package rpc

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	pb "github.com/suiteserve/protocol/go/protocol"
	"github.com/suiteserve/suiteserve/internal/repo"
	"google.golang.org/grpc"
)

type Repo interface {
	InsertAttachment(repo.Attachment) (id string, err error)
	Attachment(id string) (repo.Attachment, error)
	SuiteAttachments(suiteId string) ([]repo.Attachment, error)
	CaseAttachments(caseId string) ([]repo.Attachment, error)

	InsertSuite(repo.Suite) (id string, err error)
	Suite(id string) (repo.Suite, error)
	WatchSuites(id string, padLt, padGt int) (*repo.SuiteWatcher, error)

	InsertCase(repo.Case) (id string, err error)
	Case(id string) (repo.Case, error)

	InsertLogLine(repo.LogLine) (id string, err error)
	LogLine(id string) (repo.LogLine, error)
}

func RegisterServices(srv *grpc.Server, maxUploadMb int, repo Repo) {
	pb.RegisterQueryServiceServer(srv, &query{
		Repo: repo,
	})
	pb.RegisterSuiteServiceServer(srv, &suite{
		Repo:        repo,
		maxUploadMb: maxUploadMb,
	})
}

// millisToPb converts the given number of milliseconds since the Unix epoch
// into the well-known Protobuf Timestamp type.
func millisToPb(t int64) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: t / 1e3,
		Nanos:   int32((t % 1e3) * 1e6),
	}
}

func buildPbAttachment(a repo.Attachment,
	ownerFunc func(*pb.Attachment)) *pb.Attachment {
	pbA := pb.Attachment{
		Id:          a.Id,
		Version:     a.Version,
		Deleted:     a.Deleted,
		DeletedAt:   millisToPb(a.DeletedAt),
		Filename:    a.Filename,
		Url:         a.Url,
		ContentType: a.ContentType,
		Size:        a.Size,
		Timestamp:   millisToPb(a.Timestamp),
	}
	ownerFunc(&pbA)
	return &pbA
}

func buildPbSuite(s repo.Suite) *pb.Suite {
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
		panic("unknown suite status")
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
		panic("unknown suite result")
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
