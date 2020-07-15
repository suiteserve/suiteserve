package rpc

import (
	"context"
	pb "github.com/suiteserve/protocol/go/protocol"
)

type query struct {
	pb.UnimplementedQueryServiceServer
}

func (s *query) GetAttachments(ctx context.Context, r *pb.GetAttachmentsRequest) (*pb.GetAttachmentsReply, error) {
	panic("implement me")
}

func (s *query) WatchSuites(stream pb.QueryService_WatchSuitesServer) error {
	panic("implement me")
}

func (s *query) WatchCases(stream pb.QueryService_WatchCasesServer) error {
	panic("implement me")
}

func (s *query) WatchLog(stream pb.QueryService_WatchLogServer) error {
	panic("implement me")
}
