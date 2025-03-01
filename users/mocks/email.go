package mocks

import (
	"github.com/hantdev/viot/errors"
	"github.com/hantdev/viot/users"
)

type emailerMock struct {
}

// NewEmailer provides emailer instance for the test
func NewEmailer() users.Emailer {
	return &emailerMock{}
}

func (e *emailerMock) SendPasswordReset([]string, string, string) errors.Error {
	return nil
}
