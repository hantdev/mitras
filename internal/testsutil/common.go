package testsutil

import (
	"fmt"
	"testing"

	"github.com/hantdev/mitras/pkg/uuid"
	"github.com/stretchr/testify/require"
)

func GenerateUUID(t *testing.T) string {
	idProvider := uuid.New()
	ulid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	return ulid
}
