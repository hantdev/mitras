package users

import "github.com/hantdev/athena/errors"

// Hasher specifies an API for generating hashes of plain-text strings.
type Hasher interface {

	// Hash generates the hashed string from plain-text string.
	Hash(string) (string, errors.Error)

	// Compare compares the hashed string with plain-text string.
	Compare(string, string) errors.Error
}