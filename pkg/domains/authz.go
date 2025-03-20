package domains

import (
	"context"

	"github.com/hantdev/mitras/domains"
)

type Authorization interface {
	RetrieveEntity(ctx context.Context, id string) (domains.Domain, error)
}
