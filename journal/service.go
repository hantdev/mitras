package journal

import (
	"context"

	"github.com/hantdev/mitras"
	smqauthn "github.com/hantdev/mitras/pkg/authn"
	"github.com/hantdev/mitras/pkg/errors"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
)

type service struct {
	idProvider mitras.IDProvider
	repository Repository
}

func NewService(idp mitras.IDProvider, repository Repository) Service {
	return &service{
		idProvider: idp,
		repository: repository,
	}
}

func (svc *service) Save(ctx context.Context, journal Journal) error {
	id, err := svc.idProvider.ID()
	if err != nil {
		return err
	}
	journal.ID = id

	return svc.repository.Save(ctx, journal)
}

func (svc *service) RetrieveAll(ctx context.Context, session smqauthn.Session, page Page) (JournalsPage, error) {
	journalPage, err := svc.repository.RetrieveAll(ctx, page)
	if err != nil {
		return JournalsPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	return journalPage, nil
}
