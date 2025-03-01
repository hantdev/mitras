package users

import (
	"context"
	"regexp"
	"strings"

	"github.com/hantdev/viot/errors"
)

var (
	// Check local-part of the email address for valid characters
	userRegexp = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+/=?^_`{|}~.-]+$")
	// Check domain part (host) of the email address (example.com)
	hostRegexp    = regexp.MustCompile(`^[^\s]+\.[^\s]+$`)
	userDotRegexp = regexp.MustCompile("(^[.]{1})|([.]{1}$)|([.]{2,})")
)

// User represents a vIoT user account.
// Each user is identified by its email address and password.
type User struct {
	Email    string
	Password string
	Metadata map[string]interface{}
}

// Validate validates the user account.
// It checks the email address and password.
func (u User) Validate() errors.Error {
	if u.Email == "" || u.Password == "" {
		return ErrMalformedEntity
	}

	if !isEmail(u.Email) {
		return ErrMalformedEntity
	}
	return nil
}

func isEmail(email string) bool {
	// Check email length and boundaries
	// to avoid buffer overflow attacks
	if len(email) < 6 || len(email) > 254 {
		return false
	}

	// Check @ symbol position and presence
	// in the email address
	at := strings.LastIndex(email, "@")
	if at <= 0 || at >= len(email)-3 {
		return false
	}

	// Length of the local-part of the email address
	// should not exceed 64 characters
	user := email[:at]
	host := email[at+1:]

	if len(user) > 64 {
		return false
	}

	if userDotRegexp.MatchString(user) || !userRegexp.MatchString(user) || !hostRegexp.MatchString(host) {
		return false
	}

	return true
}

// UserRepository specifies an API for user account management.
// It is used to create, update, retrieve and delete user accounts.
type UserRepository interface {
	// Save persists the user account. A non-nil error is returned to indicate
	// operation failure.
	Save(context.Context, User) errors.Error

	// Update updates the user metadata.
	UpdateUser(context.Context, User) errors.Error

	// RetrieveByID retrieves user by its unique identifier (i.e. email).
	RetrieveByID(context.Context, string) (User, errors.Error)

	// UpdatePassword updates password for user with given email
	UpdatePassword(_ context.Context, email, password string) errors.Error
}
