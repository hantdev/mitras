package auth

type authenticateRes struct {
	id       string
	userID   string
	domainID string
}

type authorizeRes struct {
	id         string
	authorized bool
}
