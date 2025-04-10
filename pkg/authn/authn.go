package authn

import (
	"context"
)

type Session struct {
	DomainUserID string
	UserID       string
	DomainID     string
	SuperAdmin   bool
}

// Authn is mitras authentication library.
//
//go:generate mockery --name Authentication --output=./mocks --filename authn.go --quiet
type Authentication interface {
	Authenticate(ctx context.Context, token string) (Session, error)
}
