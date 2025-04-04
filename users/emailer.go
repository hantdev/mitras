package users

// Emailer wrapper around the email.
//
//go:generate mockery --name Emailer --output=./mocks --filename emailer.go --quiet --note "Mitras IoT Central users emailer interface"
type Emailer interface {
	// SendPasswordReset sends an email to the user with a link to reset the password.
	SendPasswordReset(To []string, host, user, token string) error
}
