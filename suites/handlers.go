package suites

import (
	"context"
	pb "github.com/suiteserve/protocol/go/protocol"
)

type handlers struct {
	pb.UnimplementedSuiteServiceServer
}

func (h *handlers) CreateSuite(ctx context.Context, r *pb.NewSuite) (*pb.CreateSuiteReply, error) {
	panic("implement me")
}

func (h *handlers) ReconnectSuite(ctx context.Context, r *pb.ReconnectSuiteRequest) (*pb.ReconnectSuiteReply, error) {
	panic("implement me")
}

func (h *handlers) FinishSuite(ctx context.Context, r *pb.FinishSuiteRequest) (*pb.FinishSuiteReply, error) {
	panic("implement me")
}

func (h *handlers) CreateCase(ctx context.Context, r *pb.NewCase) (*pb.CreateCaseReply, error) {
	panic("implement me")
}

func (h *handlers) StartCase(ctx context.Context, r *pb.StartCaseRequest) (*pb.StartCaseReply, error) {
	panic("implement me")
}

func (h *handlers) FinishCase(ctx context.Context, r *pb.FinishCaseRequest) (*pb.FinishCaseReply, error) {
	panic("implement me")
}

func (h *handlers) CreateLogLine(ctx context.Context, r *pb.NewLogLine) (*pb.CreateLogLineReply, error) {
	panic("implement me")
}

func (h *handlers) UploadAttachment(stream pb.SuiteService_UploadAttachmentServer) error {
	panic("implement me")
}
