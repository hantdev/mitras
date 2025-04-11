package invitations

import (
	"context"
	"time"

	"github.com/hantdev/mitras/auth"
	grpcTokenV1 "github.com/hantdev/mitras/internal/grpc/token/v1"
	"github.com/hantdev/mitras/pkg/authn"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
	mgsdk "github.com/hantdev/mitras/pkg/sdk"
)

type service struct {
	token grpcTokenV1.TokenServiceClient
	repo  Repository
	sdk   mgsdk.SDK
}

func NewService(token grpcTokenV1.TokenServiceClient, repo Repository, sdk mgsdk.SDK) Service {
	return &service{
		token: token,
		repo:  repo,
		sdk:   sdk,
	}
}

func (svc *service) SendInvitation(ctx context.Context, session authn.Session, invitation Invitation) error {
	if err := CheckRelation(invitation.Relation); err != nil {
		return err
	}

	invitation.InvitedBy = session.UserID

	joinToken, err := svc.token.Issue(ctx, &grpcTokenV1.IssueReq{UserId: session.UserID, Type: uint32(auth.InvitationKey)})
	if err != nil {
		return err
	}
	invitation.Token = joinToken.GetAccessToken()

	if invitation.Resend {
		invitation.UpdatedAt = time.Now()

		return svc.repo.UpdateToken(ctx, invitation)
	}

	invitation.CreatedAt = time.Now()

	if err := svc.repo.Create(ctx, invitation); err != nil {
		return err
	}
	return nil
}

func (svc *service) ViewInvitation(ctx context.Context, session authn.Session, userID, domainID string) (invitation Invitation, err error) {
	inv, err := svc.repo.Retrieve(ctx, userID, domainID)
	if err != nil {
		return Invitation{}, err
	}
	inv.Token = ""

	return inv, nil
}

func (svc *service) ListInvitations(ctx context.Context, session authn.Session, page Page) (invitations InvitationPage, err error) {
	ip, err := svc.repo.RetrieveAll(ctx, page)
	if err != nil {
		return InvitationPage{}, err
	}
	return ip, nil
}

func (svc *service) AcceptInvitation(ctx context.Context, session authn.Session, domainID string) error {
	inv, err := svc.repo.Retrieve(ctx, session.UserID, domainID)
	if err != nil {
		return err
	}

	if inv.UserID != session.UserID {
		return svcerr.ErrAuthorization
	}

	if !inv.ConfirmedAt.IsZero() {
		return svcerr.ErrInvitationAlreadyAccepted
	}

	if !inv.RejectedAt.IsZero() {
		return svcerr.ErrInvitationAlreadyRejected
	}

	req := mgsdk.UsersRelationRequest{
		Relation: inv.Relation,
		UserIDs:  []string{session.UserID},
	}
	if sdkerr := svc.sdk.AddUserToDomain(inv.DomainID, req, inv.Token); sdkerr != nil {
		return sdkerr
	}

	inv.ConfirmedAt = time.Now()
	inv.UpdatedAt = inv.ConfirmedAt
	return svc.repo.UpdateConfirmation(ctx, inv)
}

func (svc *service) RejectInvitation(ctx context.Context, session authn.Session, domainID string) error {
	inv, err := svc.repo.Retrieve(ctx, session.UserID, domainID)
	if err != nil {
		return err
	}

	if inv.UserID != session.UserID {
		return svcerr.ErrAuthorization
	}

	if !inv.ConfirmedAt.IsZero() {
		return svcerr.ErrInvitationAlreadyAccepted
	}

	if !inv.RejectedAt.IsZero() {
		return svcerr.ErrInvitationAlreadyRejected
	}

	inv.RejectedAt = time.Now()
	inv.UpdatedAt = inv.RejectedAt
	return svc.repo.UpdateRejection(ctx, inv)
}

func (svc *service) DeleteInvitation(ctx context.Context, session authn.Session, userID, domainID string) error {
	if session.UserID == userID {
		return svc.repo.Delete(ctx, userID, domainID)
	}

	inv, err := svc.repo.Retrieve(ctx, userID, domainID)
	if err != nil {
		return err
	}

	if inv.InvitedBy == session.UserID {
		return svc.repo.Delete(ctx, userID, domainID)
	}

	return svc.repo.Delete(ctx, userID, domainID)
}
