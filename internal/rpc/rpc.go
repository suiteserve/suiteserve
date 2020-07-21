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

func suiteToPb(s repo.Suite) *pb.Suite {
	var status pb.SuiteStatus
	switch s.Status {
	case repo.SuiteStatusUnknown:
		status = pb.SuiteStatus_SUITE_STATUS_UNSPECIFIED
	case repo.SuiteStatusStarted:
		status = pb.SuiteStatus_SUITE_STATUS_STARTED
	case repo.SuiteStatusFinished:
		status = pb.SuiteStatus_SUITE_STATUS_FINISHED
	case repo.SuiteStatusDisconnected:
		status = pb.SuiteStatus_SUITE_STATUS_DISCONNECTED
	default:
		panic("unknown status")
	}

	var result pb.SuiteResult
	switch s.Result {
	case repo.SuiteResultUnknown:
		result = pb.SuiteResult_SUITE_RESULT_UNSPECIFIED
	case repo.SuiteResultPassed:
		result = pb.SuiteResult_SUITE_RESULT_PASSED
	case repo.SuiteResultFailed:
		result = pb.SuiteResult_SUITE_RESULT_FAILED
	default:
		panic("unknown result")
	}

	return &pb.Suite{
		Id:             s.Id,
		Version:        s.Version,
		Deleted:        s.Deleted,
		DeletedAt:      millisToPb(s.DeletedAt),
		Name:           s.Name,
		Tags:           s.Tags,
		PlannedCases:   s.PlannedCases,
		Status:         status,
		Result:         result,
		DisconnectedAt: millisToPb(s.DisconnectedAt),
		StartedAt:      millisToPb(s.StartedAt),
		FinishedAt:     millisToPb(s.FinishedAt),
	}
}
