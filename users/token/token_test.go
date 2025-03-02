package token_test

import (
	"fmt"
	"testing"

	"github.com/hantdev/athena/users/token"
	"github.com/stretchr/testify/assert"
)

var email = "example@gmai.com"

func TestGenerate(t *testing.T) {
	_, err := token.New([]byte("secret"), 1).Generate(email, 0)
	assert.Nil(t, err, fmt.Sprintf("Testing token generate: %s.\n", err))
}

func TestVerify(t *testing.T) {
	tok, err := token.New([]byte("secret"), 1).Generate(email, 0)
	assert.Nil(t, err, fmt.Sprintf("Token generation failed: %s.\n", err))

	_, err = token.New([]byte("secret"), 1).Verify(tok)
	assert.Nil(t, err, fmt.Sprintf("Token verification failed: %s.\n", err))

}
