package mocks

import (
	"github.com/hantdev/athena/users"
	"github.com/hantdev/athena/users/token"
)

// NewTokenizer provides tokenizer for the test
func NewTokenizer() users.Tokenizer {
	return token.New([]byte("secret"), 1)
}
