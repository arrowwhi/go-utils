package handler

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/arrowwhi/go-utils/grpcserver/test/proto"
)

type Service struct {
	logger *zap.Logger
}

func (s *Service) GetStatusInfo(ctx context.Context, req *pb.Req) (*pb.Resp, error) {
	return &pb.Resp{
		Input:  req.Input,
		Output: 42,
	}, nil
}

func New(logger *zap.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) RegisterServer(server *grpc.Server) {
	pb.RegisterUsersServiceServer(server, s)
}

func (s *Service) RegisterHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pb.RegisterUsersServiceHandler(ctx, mux, conn)
}
