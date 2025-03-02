package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hantdev/athena/errors"
	"github.com/hantdev/athena/users"
)

const (
	issuer   string        = "athena"
	duration time.Duration = 10 * time.Hour
)

var _ users.IdentityProvider = (*jwtIdentityProvider)(nil)

type jwtIdentityProvider struct {
	secret string
}

// New instantiates a new JWT identity provider
func New(secret string) users.IdentityProvider {
	return &jwtIdentityProvider{secret}
}

// jwt generates a JWT token for the given claims
func (idp *jwtIdentityProvider) jwt(claims jwt.RegisteredClaims) (string, errors.Error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // HMAC-SHA256 algorithm
	tok, err := token.SignedString([]byte(idp.secret))
	if err != nil {
		return tok, errors.Wrap(users.ErrGetToken, err)
	}
	return tok, nil
}


// TemporaryKey generates a temporary key for the given identity
func (idp *jwtIdentityProvider) TemporaryKey(id string) (string, errors.Error) {
	now := time.Now().UTC()
	exp := now.Add(duration)

	claims := jwt.RegisteredClaims{
		Subject:   id,
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
	}
	return idp.jwt(claims)
}

// Identity retrieves the identity from the given key
func (idp *jwtIdentityProvider) Identity(key string) (string, errors.Error) {

	// Parse the token
	token, err := jwt.Parse(key, func(token *jwt.Token) (interface{}, error) {

		// Check the signing method of the token is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, users.ErrUnauthorizedAccess
		}

		return []byte(idp.secret), nil
	})

	if err != nil {
		return "", errors.Wrap(users.ErrUnauthorizedAccess, err)

	}

	// Validate the token and extract the subject from the claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if sub := claims["sub"]; sub != nil {
			return sub.(string), nil
		}
	}

	return "", users.ErrUnauthorizedAccess
}
