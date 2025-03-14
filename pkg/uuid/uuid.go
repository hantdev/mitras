// Package uuid provides a UUID identity provider.
package uuid

import (
	"github.com/gofrs/uuid/v5"
	"github.com/hantdev/mitras"
	"github.com/hantdev/mitras/pkg/errors"
)

// ErrGeneratingID indicates error in generating UUID.
var ErrGeneratingID = errors.New("failed to generate uuid")

var _ mitras.IDProvider = (*uuidProvider)(nil)

type uuidProvider struct{}

// New instantiates a UUID provider.
func New() mitras.IDProvider {
	return &uuidProvider{}
}

func (up *uuidProvider) ID() (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", errors.Wrap(ErrGeneratingID, err)
	}

	return id.String(), nil
}
