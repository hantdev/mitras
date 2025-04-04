package users

// Emailer wrapper around the email.
type Emailer interface {
	// SendPasswordReset sends an email to the user with a link to reset the password.
	SendPasswordReset(To []string, host, user, token string) error
}
