package mocks

import (
	"github.com/hantdev/athena/errors"
	"github.com/hantdev/athena/users"
)

var _ users.Hasher = (*hasherMock)(nil)

type hasherMock struct{}

// NewHasher creates "no-op" hasher for test purposes. This implementation will
// return secrets without changing them.
func NewHasher() users.Hasher {
	return &hasherMock{}
}

func (hm *hasherMock) Hash(pwd string) (string, errors.Error) {
	if pwd == "" {
		return "", users.ErrMalformedEntity
	}
	return pwd, nil
}

func (hm *hasherMock) Compare(plain, hashed string) errors.Error {
	if plain != hashed {
		return users.ErrUnauthorizedAccess
	}

	return nil
}
