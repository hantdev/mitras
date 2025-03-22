package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hantdev/mitras/pkg/errors"
)

// HealthInfo contains version endpoint response.
type HealthInfo struct {
	// Status contains service status.
	Status string `json:"status"`

	// Version contains current service version.
	Version string `json:"version"`

	// Commit represents the git hash commit.
	Commit string `json:"commit"`

	// Description contains service description.
	Description string `json:"description"`

	// BuildTime contains service build time.
	BuildTime string `json:"build_time"`
}

func (sdk mitrasSDK) Health(service string) (HealthInfo, errors.SDKError) {
	var url string
	switch service {
	case "clients":
		url = fmt.Sprintf("%s/health", sdk.clientsURL)
	case "users":
		url = fmt.Sprintf("%s/health", sdk.usersURL)
	case "certs":
		url = fmt.Sprintf("%s/health", sdk.certsURL)
	case "http-adapter":
		url = fmt.Sprintf("%s/health", sdk.httpAdapterURL)
	}

	resp, err := sdk.client.Get(url)
	if err != nil {
		return HealthInfo{}, errors.NewSDKError(err)
	}
	defer resp.Body.Close()

	if err := errors.CheckError(resp, http.StatusOK); err != nil {
		return HealthInfo{}, err
	}

	var h HealthInfo
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return HealthInfo{}, errors.NewSDKError(err)
	}

	return h, nil
}
