package rpc

import (
	"context"
	"errors"
	pb "github.com/suiteserve/protocol/go/protocol"
	"github.com/suiteserve/suiteserve/internal/repo"
	"log"
)

type query struct {
	pb.UnimplementedQueryServiceServer
	Repo
}

func (s *query) GetAttachments(_ context.Context,
	r *pb.GetAttachmentsRequest) (*pb.GetAttachmentsReply, error) {
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

func (s *query) WatchSuites(r *pb.WatchSuitesRequest,
	stream pb.QueryService_WatchSuitesServer) error {
	w, err := s.Repo.WatchSuites(r.Id, int(r.PadOlder), int(r.PadNewer))
	if err != nil {
		return err
	}
	defer w.Close()
	for {
		select {
		case changefeed := <-w.Changes():
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
				default:
					panic("unknown change type")
				}
			}
			if err := stream.Send(&reply); err != nil {
				log.Printf("send WatchSuitesReply: %v", err)
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

func (s *query) WatchCases(r *pb.WatchCasesRequest,
	stream pb.QueryService_WatchCasesServer) error {
	panic("NYI")
}

func (s *query) WatchLog(r *pb.WatchLogRequest,
	stream pb.QueryService_WatchLogServer) error {
	panic("NYI")
}
