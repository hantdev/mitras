package users

// Hasher specifies an API for generating hashes of an arbitrary textual
// content.
//
//go:generate mockery --name Hasher --output=./mocks --filename hasher.go --quiet --note "Mitras IoT Central users hasher interface"
type Hasher interface {
	// Hash generates the hashed string from plain-text.
	Hash(string) (string, error)

	// Compare compares plain-text version to the hashed one. An error should
	// indicate failed comparison.
	Compare(string, string) error
}