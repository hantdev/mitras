package domainscache

import (
	"context"

	"github.com/hantdev/mitras/domains"
	"github.com/hantdev/mitras/domains/private"
	pkgDomains "github.com/hantdev/mitras/pkg/domains"
	"github.com/hantdev/mitras/pkg/errors"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
)

type authorization struct {
	psvc private.Service
}

var _ pkgDomains.Authorization = (*authorization)(nil)

func NewAuthorization(psvc private.Service) pkgDomains.Authorization {
	return authorization{
		psvc: psvc,
	}
}

func (a authorization) RetrieveEntity(ctx context.Context, id string) (domains.Domain, error) {
	dom, err := a.psvc.RetrieveEntity(ctx, id)
	if err != nil {
		return domains.Domain{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return dom, nil
}
