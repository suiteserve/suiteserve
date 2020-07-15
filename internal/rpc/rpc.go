package rpc

import (
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb "github.com/suiteserve/protocol/go/protocol"
	"google.golang.org/grpc"
	"net/http"
)

type Service struct {
	srv   *grpc.Server
	proxy *grpcweb.WrappedGrpcServer
}

func New() *Service {
	s := Service{
		srv: grpc.NewServer(),
	}
	s.proxy = grpcweb.WrapServer(s.srv, grpcweb.WithWebsockets(true),
		grpcweb.WithWebsocketOriginFunc(func(*http.Request) bool {
			// TODO: not for production
			return true
		}))
	pb.RegisterSuiteServiceServer(s.srv, &suite{})
	pb.RegisterQueryServiceServer(s.srv, &query{})
	return &s
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.proxy.IsGrpcWebSocketRequest(r) {
		s.proxy.HandleGrpcWebsocketRequest(w, r)
	} else {
		s.srv.ServeHTTP(w, r)
	}
}

func (s *Service) Stop() {
	s.srv.GracefulStop()
}
