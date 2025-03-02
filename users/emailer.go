package users

import "github.com/hantdev/athena/errors"

// Emailer wraps the email sending functionality.
type Emailer interface {
	SendPasswordReset(To []string, host, token string) errors.Error
}