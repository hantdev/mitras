package users

import "github.com/hantdev/viot/errors"

// Emailer wraps the email sending functionality.
type Emailer interface {
	SendPasswordReset(To []string, host, token string) errors.Error
}