package users

import "github.com/hantdev/viot/errors"

// IdentityProvider specifies an API for identity management via security tokens. 
// It is used to generate temporary keys (access tokens) for the user and extract
// the entity identifier given its secret key.
type IdentityProvider interface {

	// TemporaryKey generates a temporary key (access token) for the user.
	TemporaryKey(string) (string, errors.Error)

	// Identity extracts the entity identifier given its secret key.
	Identity(string) (string, errors.Error)
}