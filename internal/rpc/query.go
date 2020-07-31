package rpc

import (
	"context"
	"errors"
	pb "github.com/suiteserve/protocol/go/protocol"
	"github.com/suiteserve/suiteserve/internal/repo"
	"io"
	"log"
)

type query struct {
	pb.UnimplementedQueryServiceServer
	Repo
}

func (s *query) GetAttachments(_ context.Context, r *pb.GetAttachmentsRequest) (*pb.GetAttachmentsReply, error) {
	var err error
	var all []repo.Attachment
	var setOwner func(*pb.Attachment)

	switch r.Filter.(type) {
	case *pb.GetAttachmentsRequest_Id:
		a, err := s.Attachment(r.GetId())
		if errors.Is(err, repo.ErrNotFound) {
			err = nil
		} else if err != nil {
			return nil, err
		} else {
			all = []repo.Attachment{a}
		}
		setOwner = func(a *pb.Attachment) {
			a.Owner = nil
		}
	case *pb.GetAttachmentsRequest_SuiteId:
		if all, err = s.SuiteAttachments(r.GetSuiteId()); err != nil {
			return nil, err
		}
		setOwner = func(a *pb.Attachment) {
			a.Owner = &pb.Attachment_SuiteId{
				SuiteId: r.GetSuiteId(),
			}
		}
	case *pb.GetAttachmentsRequest_CaseId:
		if all, err = s.CaseAttachments(r.GetCaseId()); err != nil {
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
		reply.Attachments[i] = buildPbAttachment(a, setOwner)
	}
	return &reply, nil
}

func (s *query) WatchSuites(stream pb.QueryService_WatchSuitesServer) error {
	w := s.Repo.WatchSuites()
	defer w.Close()

	go func() {
		for changefeed := range w.Changes() {
			var reply pb.WatchSuitesReply
			for _, c := range changefeed {
				switch change := c.(type) {
				case repo.SuiteAggUpdate:
					reply.TotalCount = change.TotalCount
					reply.StartedCount = change.StartedCount
				case repo.SuiteUpsert:
					reply.Upserts = append(reply.Upserts,
						&pb.WatchSuitesReply_Upsert{
							Suite: buildPbSuite(change.Suite),
						})
				case repo.SuiteDelete:
					reply.DeletedIds = append(reply.DeletedIds, change.Id)
				default:
					panic("unknown change type")
				}
			}
			if err := stream.Send(&reply); err != nil {
				log.Printf("send WatchSuitesReply: %v", err)
			}
		}
	}()

	for {
		r, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		err = w.SetQuery(r.Id, int(r.PadOlder), int(r.PadNewer))
		if err != nil {
			return err
		}
	}
}

func (s *query) WatchCases(stream pb.QueryService_WatchCasesServer) error {
	panic("implement me")
}

func (s *query) WatchLog(stream pb.QueryService_WatchLogServer) error {
	panic("implement me")
}
