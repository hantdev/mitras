package token

import (
	"context"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/hantdev/mitras/auth"
	grpcapi "github.com/hantdev/mitras/auth/api/grpc"
	grpcTokenV1 "github.com/hantdev/mitras/internal/grpc/token/v1"
)

var _ grpcTokenV1.TokenServiceServer = (*tokenGrpcServer)(nil)

type tokenGrpcServer struct {
	grpcTokenV1.UnimplementedTokenServiceServer
	issue   kitgrpc.Handler
	refresh kitgrpc.Handler
}

// NewAuthServer returns new AuthnServiceServer instance.
func NewTokenServer(svc auth.Service) grpcTokenV1.TokenServiceServer {
	return &tokenGrpcServer{
		issue: kitgrpc.NewServer(
			(issueEndpoint(svc)),
			decodeIssueRequest,
			encodeIssueResponse,
		),
		refresh: kitgrpc.NewServer(
			(refreshEndpoint(svc)),
			decodeRefreshRequest,
			encodeIssueResponse,
		),
	}
}

func (s *tokenGrpcServer) Issue(ctx context.Context, req *grpcTokenV1.IssueReq) (*grpcTokenV1.Token, error) {
	_, res, err := s.issue.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcapi.EncodeError(err)
	}
	return res.(*grpcTokenV1.Token), nil
}

func (s *tokenGrpcServer) Refresh(ctx context.Context, req *grpcTokenV1.RefreshReq) (*grpcTokenV1.Token, error) {
	_, res, err := s.refresh.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcapi.EncodeError(err)
	}
	return res.(*grpcTokenV1.Token), nil
}

func decodeIssueRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*grpcTokenV1.IssueReq)
	return issueReq{
		userID:  req.GetUserId(),
		keyType: auth.KeyType(req.GetType()),
	}, nil
}

func decodeRefreshRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*grpcTokenV1.RefreshReq)
	return refreshReq{refreshToken: req.GetRefreshToken()}, nil
}

func encodeIssueResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(issueRes)

	return &grpcTokenV1.Token{
		AccessToken:  res.accessToken,
		RefreshToken: &res.refreshToken,
		AccessType:   res.accessType,
	}, nil
}
