package mocks

import (
	"github.com/hantdev/athena/users"
	"github.com/hantdev/athena/errors"
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
