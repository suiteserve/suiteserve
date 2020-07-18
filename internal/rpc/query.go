package rpc

import (
	"context"
	"errors"
	pb "github.com/suiteserve/protocol/go/protocol"
	"github.com/suiteserve/suiteserve/internal/repo"
)

type query struct {
	pb.UnimplementedQueryServiceServer
	Repo
}

func (s *query) GetAttachments(_ context.Context, r *pb.GetAttachmentsRequest) (*pb.GetAttachmentsReply, error) {
	var err error
	var all []*repo.Attachment
	var setOwner func(*pb.Attachment)

	switch r.Filter.(type) {
	case *pb.GetAttachmentsRequest_Id:
		a, err := s.Attachment(r.GetId())
		if errors.Is(err, repo.ErrNotFound) {
			a, err = nil, nil
		} else if err != nil {
			return nil, err
		}
		if a != nil {
			all = []*repo.Attachment{a}
		}
		setOwner = func(a *pb.Attachment) {
			a.Owner = nil
		}
	case *pb.GetAttachmentsRequest_SuiteId:
		all, err = s.SuiteAttachments(r.GetSuiteId())
		if err != nil {
			return nil, err
		}
		setOwner = func(a *pb.Attachment) {
			a.Owner = &pb.Attachment_SuiteId{
				SuiteId: r.GetSuiteId(),
			}
		}
	case *pb.GetAttachmentsRequest_CaseId:
		all, err = s.CaseAttachments(r.GetCaseId())
		if err != nil {
			return nil, err
		}
		setOwner = func(a *pb.Attachment) {
			a.Owner = &pb.Attachment_CaseId{
				CaseId: r.GetCaseId(),
			}
		}
	default:
		panic("unknown filter type")
	}

	reply := pb.GetAttachmentsReply{
		Attachments: make([]*pb.Attachment, len(all)),
	}
	for i, a := range all {
		reply.Attachments[i] = &pb.Attachment{
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
		setOwner(reply.Attachments[i])
	}
	return &reply, nil
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
