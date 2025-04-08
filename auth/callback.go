package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hantdev/mitras/pkg/errors"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
	"github.com/hantdev/mitras/pkg/policies"
)

type callback struct {
	httpClient        *http.Client
	urls              []string
	method            string
	allowedPermission map[string]struct{}
}

// CallBack send auth request to an external service.
type CallBack interface {
	Authorize(ctx context.Context, pr policies.Policy) error
}

// NewCallback creates a new instance of CallBack.
func NewCallback(httpClient *http.Client, method string, urls []string, permissions []string) (CallBack, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if method != http.MethodPost && method != http.MethodGet {
		return nil, fmt.Errorf("unsupported auth callback method: %s", method)
	}

	allowedPermission := make(map[string]struct{})
	for _, permission := range permissions {
		allowedPermission[permission] = struct{}{}
	}

	return &callback{
		httpClient:        httpClient,
		urls:              urls,
		method:            method,
		allowedPermission: allowedPermission,
	}, nil
}

func (c *callback) Authorize(ctx context.Context, pr policies.Policy) error {
	if len(c.urls) == 0 {
		return nil
	}

	// Check if the permission is in the allowed list
	// Otherwise, only call webhook if the permission is in the map
	if len(c.allowedPermission) > 0 {
		_, exists := c.allowedPermission[pr.Permission]
		if !exists {
			return nil
		}
	}

	payload := map[string]string{
		"domain":           pr.Domain,
		"subject":          pr.Subject,
		"subject_type":     pr.SubjectType,
		"subject_kind":     pr.SubjectKind,
		"subject_relation": pr.SubjectRelation,
		"object":           pr.Object,
		"object_type":      pr.ObjectType,
		"object_kind":      pr.ObjectKind,
		"relation":         pr.Relation,
		"permission":       pr.Permission,
	}

	var err error
	// We iterate through all URLs in sequence
	// if any request fails, we return the error immediately
	for _, url := range c.urls {
		if err = c.makeRequest(ctx, url, payload); err != nil {
			return err
		}
	}

	return nil
}

func (c *callback) makeRequest(ctx context.Context, urlStr string, params map[string]string) error {
	var req *http.Request
	var err error

	switch c.method {
	case http.MethodGet:
		query := url.Values{}
		for key, value := range params {
			query.Set(key, value)
		}
		req, err = http.NewRequestWithContext(ctx, c.method, urlStr+"?"+query.Encode(), nil)
	case http.MethodPost:
		data, jsonErr := json.Marshal(params)
		if jsonErr != nil {
			return jsonErr
		}
		req, err = http.NewRequestWithContext(ctx, c.method, urlStr, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
	}

	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, resp.StatusCode)
	}

	return nil
}
