package rpc

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb "github.com/suiteserve/protocol/go/protocol"
	"github.com/suiteserve/suiteserve/internal/repo"
	"google.golang.org/grpc"
	"net/http"
	"time"
)

type Repo interface {
	Attachment(id string) (*repo.Attachment, error)
	SuiteAttachments(suiteId string) ([]*repo.Attachment, error)
	CaseAttachments(caseId string) ([]*repo.Attachment, error)
}

type Service struct {
	srv         *grpc.Server
	proxy       *grpcweb.WrappedGrpcServer
	maxUploadMb int
	repo        Repo
}

func New(maxUploadMb int, repo Repo) *Service {
	s := Service{
		srv:         grpc.NewServer(),
		maxUploadMb: maxUploadMb,
		repo:        repo,
	}
	s.proxy = grpcweb.WrapServer(s.srv, grpcweb.WithWebsockets(true),
		grpcweb.WithWebsocketOriginFunc(func(*http.Request) bool {
			// TODO: not for production
			return true
		}))
	pb.RegisterSuiteServiceServer(s.srv, &suite{Service: &s})
	pb.RegisterQueryServiceServer(s.srv, &query{Service: &s})
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

func timeToPb(t time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}
