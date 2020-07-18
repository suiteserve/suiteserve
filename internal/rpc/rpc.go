package rpc

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb "github.com/suiteserve/protocol/go/protocol"
	"github.com/suiteserve/suiteserve/event"
	"github.com/suiteserve/suiteserve/internal/repo"
	"google.golang.org/grpc"
	"net/http"
)

type Repo interface {
	Changefeed() *event.Bus

	InsertAttachment(repo.Attachment) (id string, err error)
	Attachment(id string) (*repo.Attachment, error)
	SuiteAttachments(suiteId string) ([]*repo.Attachment, error)
	CaseAttachments(caseId string) ([]*repo.Attachment, error)

	InsertSuite(repo.Suite) (id string, err error)
	Suite(id string) (*repo.Suite, error)

	InsertCase(repo.Case) (id string, err error)
	Case(id string) (*repo.Case, error)

	InsertLogLine(repo.LogLine) (id string, err error)
	LogLine(id string) (*repo.LogLine, error)
}

type Service struct {
	srv   *grpc.Server
	proxy *grpcweb.WrappedGrpcServer
	repo  Repo
}

func New(maxUploadMb int, repo Repo) *Service {
	s := Service{
		srv:  grpc.NewServer(),
		repo: repo,
	}
	s.proxy = grpcweb.WrapServer(s.srv, grpcweb.WithWebsockets(true),
		grpcweb.WithWebsocketOriginFunc(func(*http.Request) bool {
			// TODO: not for production
			return true
		}))
	pb.RegisterQueryServiceServer(s.srv, &query{
		Repo: repo,
	})
	pb.RegisterSuiteServiceServer(s.srv, &suite{
		Repo:        repo,
		maxUploadMb: maxUploadMb,
	})
	return &s
}

func (s *Service) NewMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.proxy.IsGrpcWebSocketRequest(r) {
			s.proxy.HandleGrpcWebsocketRequest(w, r)
		} else if s.proxy.IsGrpcWebRequest(r) {
			s.srv.ServeHTTP(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func (s *Service) Stop() {
	s.srv.GracefulStop()
}

// millisToPb converts the given number of milliseconds since the Unix epoch
// into the well-known Protobuf Timestamp type.
func millisToPb(t int64) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: t / 1e3,
		Nanos:   int32((t % 1e3) * 1e6),
	}
}
