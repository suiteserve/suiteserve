package rpc

import (
	"context"
	"errors"
	pb "github.com/suiteserve/protocol/go/protocol"
	"github.com/suiteserve/suiteserve/internal/repo"
	"io"
	"sort"
	"sync"
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

type suiteHandle struct {
	startedAt int64
	id        string
}

type suitesWatch struct {
	mu   sync.Mutex
	repo Repo
	id   string

	minSize int
	window  []suiteHandle
}

func (w *suitesWatch) process(r *pb.WatchSuitesRequest) ([]*pb.WatchSuitesReply, error) {
	var replies []*pb.WatchSuitesReply
	w.mu.Lock()
	defer w.mu.Unlock()
	return replies, nil
}

func (w *suitesWatch) onInsert(u repo.SuiteInsert) (*pb.WatchSuitesReply, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	handle := suiteHandle{
		startedAt: u.Suite.StartedAt,
		id:        u.Suite.Id,
	}
	if len(w.window) == 0 {
		// expand window
		w.window = []suiteHandle{handle}
	} else if u.Suite.StartedAt >= w.window[0].startedAt {
		// expand window
		var ok bool
		for i, h := range w.window {
			if u.Suite.StartedAt <= h.startedAt {
				w.window = append(w.window, suiteHandle{})
				copy(w.window[i+1:], w.window[i:])
				w.window[i] = handle
				ok = true
				break
			}
		}
		if !ok {
			w.window = append(w.window, handle)
		}
		// shrink window if possible
		left := w.findBestLeft()
		if len(w.window)-left >= w.minSize {
			w.window = w.window[left:]
		}
	} else {
		// out of range
		return nil, nil
	}
	return &pb.WatchSuitesReply{
		Operation: &pb.WatchSuitesReply_Update_{
			Update: &pb.WatchSuitesReply_Update{
				Suite: suiteToPb(u.Suite),
			},
		},
		TotalCount:   u.Agg.TotalCount,
		StartedCount: u.Agg.StartedCount,
		HasMore:      false,
	}, nil
}

func (w *suitesWatch) findBestLeft() int {
	var v int64
	for i, h := range w.window {
		// TODO: search through
		if i == 0 {
			v = h.startedAt
		} else if h.startedAt != v {
			return i
		}
	}
	return len(w.window)
}

func (w *suitesWatch) onUpdate(u repo.SuiteUpdate) *pb.WatchSuitesReply {
	w.mu.Lock()
	defer w.mu.Unlock()
	return nil
}

func (s *query) WatchSuites(stream pb.QueryService_WatchSuitesServer) error {
	sub := s.Changefeed().Subscribe()
	defer sub.Unsubscribe()

	r, err := stream.Recv()
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	_ = r

	go func() {
		for {
			changefeed, ok := (<-sub.Ch()).(repo.Changefeed)
			if !ok {
				return
			}
			for _, c := range changefeed {
				_ = c
				// TODO
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
		_ = r

		// TODO

		// stream.Send(&pb.WatchSuitesReply{
		// 	Operation:    nil,
		// 	HasMore:      false,
		// 	TotalCount:   0,
		// 	RunningCount: 0,
		// })
	}
}

func (s *query) WatchCases(stream pb.QueryService_WatchCasesServer) error {
	panic("implement me")
}

func (s *query) WatchLog(stream pb.QueryService_WatchLogServer) error {
	panic("implement me")
}

type watchHandle struct {
	id  string
	val int64
}

// less returns whether h is less than that for sorting and searching purposes.
func (h watchHandle) less(that watchHandle) bool {
	return h.val < that.val
}

type watchWindow struct {
	mu       sync.Mutex
	required *watchHandle
	minSize  int
	window   []watchHandle
}

// insert inserts h into the sorted window if any of:
//   len(window) < minSize
//   !h.less(window[0])
//
// If the insertion is not made, removed will be empty and ok will be false.
// Otherwise, the window is shrunk after the insertion and the removed
// watchHandles due to the shrinkage are returned along with ok being true.
func (w *watchWindow) insert(h watchHandle) (removed []watchHandle, ok bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.insertHelper(h)
}

func (w *watchWindow) insertHelper(h watchHandle) (removed []watchHandle, ok bool) {
	i := sort.Search(len(w.window), func(i int) bool {
		return !w.window[i].less(h)
	})
	// check
	if i == 0 && len(w.window) >= w.minSize && len(w.window) > 0 && h.less(w.window[0]) {
		return nil, false
	}
	// insert
	w.window = append(w.window, watchHandle{})
	copy(w.window[i+1:], w.window[i:])
	w.window[i] = h
	// shrink
	i = w.shrink()
	removed = make([]watchHandle, i)
	copy(removed, w.window[:i])
	w.window = w.window[i:]
	return removed, true
}

// update immediately removes h from the window if possible and then reinserts
// it according to insert. h will still have been removed even when ok is false
// and h will never show up in removed.
func (w *watchWindow) update(h watchHandle) (removed []watchHandle, ok bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i, that := range w.window {
		if that.id == h.id {
			w.window = append(w.window[:i], w.window[i+1:]...)
			break
		}
	}
	return w.insertHelper(h)
}

// shrink returns the largest left bound i of the window such that:
//   if len(window) >= minSize, then len(window[i:]) >= minSize
//   if required is non-nil, then window[i:] contains required
//   if i > 0, then window[i-1].less(window[i])
// Preconditions:
//   window is in ascending order
//   if required is non-nil, then window contains required
// If minSize is non-positive, or if len(window) <= minSize, shrink returns
// 0 to indicate that the window should not be shrunk.
func (w *watchWindow) shrink() int {
	if w.minSize < 1 || len(w.window) <= w.minSize {
		// don't shrink
		return 0
	}
	j := len(w.window) - w.minSize
	target := w.window[j]
	if w.required != nil && w.required.less(target) {
		target = *w.required
	}
	return sort.Search(j, func(i int) bool {
		return !w.window[i].less(target)
	})
}
