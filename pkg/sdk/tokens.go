package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hantdev/mitras/pkg/errors"
)

// Token is used for authentication purposes.
// It contains AccessToken, RefreshToken and AccessExpiry.
type Token struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	AccessType   string `json:"access_type,omitempty"`
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (sdk mgSDK) CreateToken(ctx context.Context, lt Login) (Token, errors.SDKError) {
	data, err := json.Marshal(lt)
	if err != nil {
		return Token{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.usersURL, usersEndpoint, issueTokenEndpoint)

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodPost, url, "", data, nil, http.StatusCreated)
	if sdkerr != nil {
		return Token{}, sdkerr
	}
	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return Token{}, errors.NewSDKError(err)
	}

	return token, nil
}

func (sdk mgSDK) RefreshToken(ctx context.Context, token string) (Token, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s", sdk.usersURL, usersEndpoint, refreshTokenEndpoint)

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodPost, url, token, nil, nil, http.StatusCreated)
	if sdkerr != nil {
		return Token{}, sdkerr
	}

	t := Token{}
	if err := json.Unmarshal(body, &t); err != nil {
		return Token{}, errors.NewSDKError(err)
	}

	return t, nil
}
