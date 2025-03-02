package mocks

import (
	"github.com/hantdev/athena/users"
	"github.com/hantdev/athena/users/jwt"
)

// NewIdentityProvider creates "mirror" identity provider, i.e. generated
// token will hold value provided by the caller.
func NewIdentityProvider() users.IdentityProvider {
	return jwt.New("secret")
}
