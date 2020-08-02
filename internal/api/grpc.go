package api

import (
	"context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
)

type GrpcService struct {
	srv *grpc.Server
}

func NewGrpcService(srv *grpc.Server) *GrpcService {
	return &GrpcService{srv: srv}
}

func (s *GrpcService) Serve(ctx context.Context, ln net.Listener) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return s.srv.Serve(ln)
	})
	eg.Go(func() error {
		<-ctx.Done()
		s.srv.GracefulStop()
		return nil
	})
	return eg.Wait()
}
