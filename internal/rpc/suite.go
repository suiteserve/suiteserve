package rpc

import (
	"context"
	pb "github.com/suiteserve/protocol/go/protocol"
)

type suite struct {
	pb.UnimplementedSuiteServiceServer
	Repo

	maxUploadMb int
}

func (s *suite) CreateSuite(ctx context.Context,
	r *pb.CreateSuiteRequest) (*pb.CreateSuiteReply, error) {
	panic("NYI")
}

func (s *suite) ReconnectSuite(ctx context.Context,
	r *pb.ReconnectSuiteRequest) (*pb.ReconnectSuiteReply, error) {
	panic("NYI")
}

func (s *suite) FinishSuite(ctx context.Context,
	r *pb.FinishSuiteRequest) (*pb.FinishSuiteReply, error) {
	panic("NYI")
}

func (s *suite) CreateCase(ctx context.Context,
	r *pb.CreateCaseRequest) (*pb.CreateCaseReply, error) {
	panic("NYI")
}

func (s *suite) StartCase(ctx context.Context,
	r *pb.StartCaseRequest) (*pb.StartCaseReply, error) {
	panic("NYI")
}

func (s *suite) FinishCase(ctx context.Context,
	r *pb.FinishCaseRequest) (*pb.FinishCaseReply, error) {
	panic("NYI")
}

func (s *suite) CreateLogLine(ctx context.Context,
	r *pb.CreateLogLineRequest) (*pb.CreateLogLineReply, error) {
	panic("NYI")
}

func (s *suite) UploadAttachment(
	stream pb.SuiteService_UploadAttachmentServer) error {
	panic("NYI")
}
