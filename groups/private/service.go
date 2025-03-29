package private

import (
	"context"

	"github.com/hantdev/mitras/groups"
)

//go:generate mockery --name Service  --output=./mocks --filename service.go --quiet --note "Service is a service for groups"
type Service interface {
	RetrieveById(ctx context.Context, id string) (groups.Group, error)
}

var _ Service = (*service)(nil)

func New(repo groups.Repository) Service {
	return service{repo}
}

type service struct {
	repo groups.Repository
}

func (svc service) RetrieveById(ctx context.Context, ids string) (groups.Group, error) {
	return svc.repo.RetrieveByID(ctx, ids)
}
