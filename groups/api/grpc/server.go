package grpc

import (
	"context"

	grpcCommonV1 "github.com/hantdev/mitras/api/grpc/common/v1"
	grpcGroupsV1 "github.com/hantdev/mitras/api/grpc/groups/v1"
	apiutil "github.com/hantdev/mitras/api/http/util"
	mitrasauth "github.com/hantdev/mitras/auth"
	groups "github.com/hantdev/mitras/groups/private"
	"github.com/hantdev/mitras/pkg/errors"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ grpcGroupsV1.GroupsServiceServer = (*grpcServer)(nil)

type grpcServer struct {
	grpcGroupsV1.UnimplementedGroupsServiceServer
	retrieveEntity kitgrpc.Handler
}

// NewServer returns new AuthServiceServer instance.
func NewServer(svc groups.Service) grpcGroupsV1.GroupsServiceServer {
	return &grpcServer{
		retrieveEntity: kitgrpc.NewServer(
			retrieveEntityEndpoint(svc),
			decodeRetrieveEntityRequest,
			encodeRetrieveEntityResponse,
		),
	}
}

func (s *grpcServer) RetrieveEntity(ctx context.Context, req *grpcCommonV1.RetrieveEntityReq) (*grpcCommonV1.RetrieveEntityRes, error) {
	_, res, err := s.retrieveEntity.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*grpcCommonV1.RetrieveEntityRes), nil
}

func decodeRetrieveEntityRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*grpcCommonV1.RetrieveEntityReq)
	return retrieveEntityReq{
		Id: req.GetId(),
	}, nil
}

func encodeRetrieveEntityResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(retrieveEntityRes)

	return &grpcCommonV1.RetrieveEntityRes{
		Entity: &grpcCommonV1.EntityBasic{
			Id:            res.id,
			DomainId:      res.domain,
			ParentGroupId: res.parentGroup,
			Status:        uint32(res.status),
		},
	}, nil
}

func encodeError(err error) error {
	switch {
	case errors.Contains(err, nil):
		return nil
	case errors.Contains(err, errors.ErrMalformedEntity),
		err == apiutil.ErrInvalidAuthKey,
		err == apiutil.ErrMissingID,
		err == apiutil.ErrMissingMemberType,
		err == apiutil.ErrMissingPolicySub,
		err == apiutil.ErrMissingPolicyObj,
		err == apiutil.ErrMalformedPolicyAct:
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Contains(err, svcerr.ErrAuthentication),
		errors.Contains(err, mitrasauth.ErrKeyExpired),
		err == apiutil.ErrMissingEmail,
		err == apiutil.ErrBearerToken:
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Contains(err, svcerr.ErrAuthorization):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}