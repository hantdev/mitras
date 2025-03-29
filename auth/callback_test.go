package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hantdev/mitras/auth"
	"github.com/hantdev/mitras/pkg/errors"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
	"github.com/hantdev/mitras/pkg/policies"
	"github.com/stretchr/testify/assert"
)

func TestCallback_Authorize(t *testing.T) {
	policy := policies.Policy{
		Domain:          "test-domain",
		Subject:         "test-subject",
		SubjectType:     "user",
		SubjectKind:     "individual",
		SubjectRelation: "owner",
		Object:          "test-object",
		ObjectType:      "message",
		ObjectKind:      "event",
		Relation:        "publish",
		Permission:      "allow",
	}

	cases := []struct {
		desc        string
		method      string
		respStatus  int
		expectError bool
	}{
		{
			desc:        "successful GET authorization",
			method:      http.MethodGet,
			respStatus:  http.StatusOK,
			expectError: false,
		},
		{
			desc:        "successful POST authorization",
			method:      http.MethodPost,
			respStatus:  http.StatusOK,
			expectError: false,
		},
		{
			desc:        "failed authorization",
			method:      http.MethodPost,
			respStatus:  http.StatusForbidden,
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tc.method, r.Method)

				if tc.method == http.MethodGet {
					query := r.URL.Query()
					assert.Equal(t, policy.Domain, query.Get("domain"))
					assert.Equal(t, policy.Subject, query.Get("subject"))
				}

				w.WriteHeader(tc.respStatus)
			}))
			defer ts.Close()

			cb, err := auth.NewCallback(http.DefaultClient, tc.method, []string{ts.URL}, []string{})
			assert.NoError(t, err)
			err = cb.Authorize(context.Background(), policy)

			if tc.expectError {
				assert.Error(t, err)
				assert.True(t, errors.Contains(err, svcerr.ErrAuthorization), "expected authorization error")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCallback_MultipleURLs(t *testing.T) {
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts1.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts2.Close()

	cb, err := auth.NewCallback(http.DefaultClient, http.MethodPost, []string{ts1.URL, ts2.URL}, []string{})
	assert.NoError(t, err)
	err = cb.Authorize(context.Background(), policies.Policy{})
	assert.NoError(t, err)
}

func TestCallback_InvalidURL(t *testing.T) {
	cb, err := auth.NewCallback(http.DefaultClient, http.MethodPost, []string{"http://invalid-url"}, []string{})
	assert.NoError(t, err)
	err = cb.Authorize(context.Background(), policies.Policy{})
	assert.Error(t, err)
}

func TestCallback_InvalidMethod(t *testing.T) {
	_, err := auth.NewCallback(http.DefaultClient, "invalid-method", []string{"http://example.com"}, []string{})
	assert.Error(t, err)
}

func TestCallback_CancelledContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cb, err := auth.NewCallback(http.DefaultClient, http.MethodPost, []string{ts.URL}, []string{})
	assert.NoError(t, err)
	err = cb.Authorize(ctx, policies.Policy{})
	assert.Error(t, err)
}

func TestNewCallback_NilClient(t *testing.T) {
	cb, err := auth.NewCallback(nil, http.MethodPost, []string{"test"}, []string{})
	assert.NoError(t, err)
	assert.NotNil(t, cb)
}

func TestCallback_NoURL(t *testing.T) {
	cb, err := auth.NewCallback(http.DefaultClient, http.MethodPost, []string{}, []string{})
	assert.NoError(t, err)
	err = cb.Authorize(context.Background(), policies.Policy{})
	assert.NoError(t, err)
}

func TestCallback_PermissionFiltering(t *testing.T) {
	webhookCalled := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	t.Run("allowed permission", func(t *testing.T) {
		webhookCalled = false
		allowedPermissions := []string{"create_client", "delete_channel"}

		cb, err := auth.NewCallback(http.DefaultClient, http.MethodPost, []string{ts.URL}, allowedPermissions)
		assert.NoError(t, err)

		err = cb.Authorize(context.Background(), policies.Policy{
			Permission: "create_client",
		})
		assert.NoError(t, err)
		assert.True(t, webhookCalled, "webhook should be called for allowed permission")
	})

	t.Run("non-allowed permission", func(t *testing.T) {
		webhookCalled = false
		allowedPermissions := []string{"create_client", "delete_channel"}

		cb, err := auth.NewCallback(http.DefaultClient, http.MethodPost, []string{ts.URL}, allowedPermissions)
		assert.NoError(t, err)

		err = cb.Authorize(context.Background(), policies.Policy{
			Permission: "read_channel",
		})
		assert.NoError(t, err)
		assert.False(t, webhookCalled, "webhook should not be called for non-allowed permission")
	})

	t.Run("empty allowed permissions", func(t *testing.T) {
		webhookCalled = false
		allowedPermissions := []string{}

		cb, err := auth.NewCallback(http.DefaultClient, http.MethodPost, []string{ts.URL}, allowedPermissions)
		assert.NoError(t, err)

		err = cb.Authorize(context.Background(), policies.Policy{
			Permission: "any_permission",
		})
		assert.NoError(t, err)
		assert.True(t, webhookCalled, "webhook should be called when allowed permissions list is empty")
	})
}
