package domains

import (
	"context"
	"time"

	"github.com/hantdev/mitras"
	"github.com/hantdev/mitras/pkg/authn"
	"github.com/hantdev/mitras/pkg/errors"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
	"github.com/hantdev/mitras/pkg/policies"
	"github.com/hantdev/mitras/pkg/roles"
)

const defLimit = 100

var (
	errCreateDomainPolicy = errors.New("failed to create domain policy")
	errRollbackRepo       = errors.New("failed to rollback repo")
)

type service struct {
	repo       Repository
	policy     policies.Service
	idProvider mitras.IDProvider
	roles.ProvisionManageService
}

var _ Service = (*service)(nil)

func New(repo Repository, policy policies.Service, idProvider mitras.IDProvider, sidProvider mitras.IDProvider, availableActions []roles.Action, builtInRoles map[roles.BuiltInRoleName][]roles.Action) (Service, error) {
	rpms, err := roles.NewProvisionManageService(policies.DomainType, repo, policy, sidProvider, availableActions, builtInRoles)
	if err != nil {
		return nil, err
	}

	return &service{
		repo:                   repo,
		policy:                 policy,
		idProvider:             idProvider,
		ProvisionManageService: rpms,
	}, nil
}

func (svc service) CreateDomain(ctx context.Context, session authn.Session, d Domain) (do Domain, err error) {
	d.CreatedBy = session.UserID

	domainID, err := svc.idProvider.ID()
	if err != nil {
		return Domain{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	d.ID = domainID

	if d.Status != DisabledStatus && d.Status != EnabledStatus {
		return Domain{}, svcerr.ErrInvalidStatus
	}

	d.CreatedAt = time.Now()

	// Domain is created in repo first, because Roles table have foreign key relation with Domain ID
	dom, err := svc.repo.Save(ctx, d)
	if err != nil {
		return Domain{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	defer func() {
		if err != nil {
			if errRollBack := svc.repo.Delete(ctx, domainID); errRollBack != nil {
				err = errors.Wrap(err, errors.Wrap(errRollbackRepo, errRollBack))
			}
		}
	}()

	newBuiltInRoleMembers := map[roles.BuiltInRoleName][]roles.Member{
		BuiltInRoleAdmin: {roles.Member(session.UserID)},
	}

	optionalPolicies := []policies.Policy{
		{
			Subject:     policies.MitrasObject,
			SubjectType: policies.PlatformType,
			Relation:    "organization",
			Object:      d.ID,
			ObjectType:  policies.DomainType,
		},
	}

	if _, err := svc.AddNewEntitiesRoles(ctx, domainID, session.UserID, []string{domainID}, optionalPolicies, newBuiltInRoleMembers); err != nil {
		return Domain{}, errors.Wrap(errCreateDomainPolicy, err)
	}

	return dom, nil
}

func (svc service) RetrieveDomain(ctx context.Context, session authn.Session, id string) (Domain, error) {
	var domain Domain
	var err error
	switch session.SuperAdmin {
	case true:
		domain, err = svc.repo.RetrieveByID(ctx, id)
	default:
		domain, err = svc.repo.RetrieveByUserAndID(ctx, session.UserID, id)
	}
	if err != nil {
		return Domain{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return domain, nil
}

func (svc service) UpdateDomain(ctx context.Context, session authn.Session, id string, d DomainReq) (Domain, error) {
	dom, err := svc.repo.Update(ctx, id, session.UserID, d)
	if err != nil {
		return Domain{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return dom, nil
}

func (svc service) EnableDomain(ctx context.Context, session authn.Session, id string) (Domain, error) {
	status := EnabledStatus
	dom, err := svc.repo.Update(ctx, id, session.UserID, DomainReq{Status: &status})
	if err != nil {
		return Domain{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return dom, nil
}

func (svc service) DisableDomain(ctx context.Context, session authn.Session, id string) (Domain, error) {
	status := DisabledStatus
	dom, err := svc.repo.Update(ctx, id, session.UserID, DomainReq{Status: &status})
	if err != nil {
		return Domain{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return dom, nil
}

// Only SuperAdmin can freeze the domain.
func (svc service) FreezeDomain(ctx context.Context, session authn.Session, id string) (Domain, error) {
	status := FreezeStatus
	dom, err := svc.repo.Update(ctx, id, session.UserID, DomainReq{Status: &status})
	if err != nil {
		return Domain{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return dom, nil
}

func (svc service) ListDomains(ctx context.Context, session authn.Session, p Page) (DomainsPage, error) {
	p.UserID = session.UserID
	if session.SuperAdmin {
		p.UserID = ""
	}

	dp, err := svc.repo.ListDomains(ctx, p)
	if err != nil {
		return DomainsPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return dp, nil
}

func (svc service) DeleteUserFromDomains(ctx context.Context, id string) (err error) {
	domainsPage, err := svc.repo.ListDomains(ctx, Page{UserID: id, Limit: defLimit})
	if err != nil {
		return err
	}

	if domainsPage.Total > defLimit {
		for i := defLimit; i < int(domainsPage.Total); i += defLimit {
			page := Page{UserID: id, Offset: uint64(i), Limit: defLimit}
			dp, err := svc.repo.ListDomains(ctx, page)
			if err != nil {
				return err
			}
			domainsPage.Domains = append(domainsPage.Domains, dp.Domains...)
		}
	}

	// if err := svc.RemoveMembersFromAllRoles(ctx, authn.Session{}, []string{id}); err != nil {
	// 	return err
	// }
	////////////ToDo//////////////
	// Remove user from all roles in all domains
	//////////////////////////

	// for _, domain := range domainsPage.Domains {
	// 	req := policies.Policy{
	// 		Subject:     policies.EncodeDomainUserID(domain.ID, id),
	// 		SubjectType: policies.UserType,
	// 	}
	// 	if err := svc.policies.DeletePolicyFilter(ctx, req); err != nil {
	// 		return err
	// 	}
	// }

	// if err := svc.repo.DeleteUserPolicies(ctx, id); err != nil {
	// 	return err
	// }

	return nil
}
