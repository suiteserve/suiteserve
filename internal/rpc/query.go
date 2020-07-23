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

type suiteWatchWindow struct {
	mu    sync.Mutex
	inner watchWindow
	repo  Repo
}

func (w *suiteWatchWindow) process(r *pb.WatchSuitesRequest) (*pb.WatchSuitesReply, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	var ok bool
	var combined pb.WatchSuitesReply
	if int(r.Size) != w.inner.minSize {
		reply, err := w.processSize(int(r.Size))
		if err != nil {
			return nil, err
		}
		ok = true
		mergeWatchSuitesReply(&combined, reply)
	}
	if r.Id != w.inner.required.getId() && !w.inner.hasId(r.Id) {
		reply, err := w.processId(r.Id)
		if err != nil {
			return nil, err
		}
		ok = true
		mergeWatchSuitesReply(&combined, reply)
	}
	if !ok {
		return nil, nil
	}
	return &combined, nil
}

func (w *suiteWatchWindow) processSize(size int) (*pb.WatchSuitesReply, error) {
	var reply pb.WatchSuitesReply
	removed := w.inner.setMinSize(size)
	reply.DeletedIds = append(reply.DeletedIds, removed...)

	if n := size - w.inner.len(); n > 0 {
		page, err := w.repo.SuitePage(w.inner.minId(), n)
		if err != nil {
			return nil, err
		}
		for _, s := range page.Suites {
			reply.Updates = append(reply.Updates, &pb.WatchSuitesReply_Update{
				Suite: buildPbSuite(s),
			})
		}
		addSuiteAgg(page.SuiteAgg, &reply)
	}
	return &reply, nil
}

func (w *suiteWatchWindow) processId(id string) (*pb.WatchSuitesReply, error) {
	// TODO
	panic("NYI")
}

func (w *suiteWatchWindow) insert(c repo.SuiteInsert) *pb.WatchSuitesReply {
	w.mu.Lock()
	defer w.mu.Unlock()
	ok := w.inner.insert(watchHandle{
		id:  c.Suite.Id,
		val: c.Suite.StartedAt,
	})
	if !ok {
		return nil
	}
	return buildPbWatchSuitesReply(c.Suite, nil, c.Agg, w.inner.shrink())
}

func (w *suiteWatchWindow) update(c repo.SuiteUpdate) *pb.WatchSuitesReply {
	w.mu.Lock()
	defer w.mu.Unlock()
	ok := w.inner.update(watchHandle{
		id:  c.Suite.Id,
		val: c.Suite.StartedAt,
	})
	if !ok {
		return nil
	}
	return buildPbWatchSuitesReply(c.Suite, c.Mask, c.Agg, w.inner.shrink())
}

func (s *query) WatchSuites(stream pb.QueryService_WatchSuitesServer) error {
	sub := s.Changefeed().Subscribe()
	defer sub.Unsubscribe()

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

func (h *watchHandle) getId() string {
	if h == nil {
		return ""
	}
	return h.id
}

// less returns whether h is less than that for sorting and searching purposes.
func (h watchHandle) less(that watchHandle) bool {
	return h.val < that.val
}

type watchWindow struct {
	required *watchHandle
	minSize  int
	window   []watchHandle
}

// len returns the length of the window.
func (w *watchWindow) len() int {
	return len(w.window)
}

// setMinSize sets minSize to a new value; if the new minSize is less than the
// old minSize, then shrink is called and its result is returned, else shrink is
// not called and the returned slice is nil.
func (w *watchWindow) setMinSize(minSize int) []string {
	if minSize < w.minSize {
		w.minSize = minSize
		return w.shrink()
	}
	w.minSize = minSize
	return nil
}

func (w *watchWindow) hasId(id string) bool {
	for _, h := range w.window {
		if h.id == id {
			return true
		}
	}
	return false
}

// minId returns the ID of the lowest watchHandle in the window, or the empty
// string if the window is empty.
func (w *watchWindow) minId() string {
	if len(w.window) > 0 {
		return w.window[0].id
	}
	return ""
}

// insert inserts h into the sorted window. If the length of the window is less
// than minSize, then h is always inserted. Otherwise, h is not inserted iff it
// comes strictly before the first element of the window. insert returns true
// iff h was successfully inserted.
func (w *watchWindow) insert(h watchHandle) bool {
	i := sort.Search(len(w.window), func(i int) bool {
		return !w.window[i].less(h)
	})
	// check
	if i == 0 && len(w.window) >= w.minSize && len(w.window) > 0 && h.less(w.window[0]) {
		return false
	}
	// insert
	w.window = append(w.window, watchHandle{})
	copy(w.window[i+1:], w.window[i:])
	w.window[i] = h
	return true
}

// update immediately removes h from the window if it exists and then reinserts
// it according to insert. update returns true iff h was successfully
// reinserted.
func (w *watchWindow) update(h watchHandle) bool {
	for i, that := range w.window {
		if that.id == h.id {
			w.window = append(w.window[:i], w.window[i+1:]...)
			break
		}
	}
	return w.insert(h)
}

// shrink sets the window to a slice of itself (window = window[i:]) where i is
// the largest index such that the new length is still at least minSize and that
// the new window still contains the required watchHandle if non-nil. In
// addition, consecutive window elements will not be split. That is, when:
//   required = nil
//   minSize = 2
//   window = [3, 4, 4, 5]
// shrink will set the window to window[1:], not window[2:], because the latter
// would split the block of consecutive 4's.
//
// If minSize is non-positive, or if the length of the window is not already at
// least minSize, then shrink does nothing. The returned string slice contains
// the watchHandle ids removed due to shrinkage.
func (w *watchWindow) shrink() []string {
	if w.minSize < 1 || len(w.window) <= w.minSize {
		// don't shrink
		return nil
	}
	j := len(w.window) - w.minSize
	target := w.window[j]
	if w.required != nil && w.required.less(target) {
		target = *w.required
	}
	i := sort.Search(j, func(i int) bool {
		return !w.window[i].less(target)
	})
	removed := make([]string, i)
	for i, h := range w.window[:i] {
		removed[i] = h.id
	}
	w.window = w.window[i:]
	return removed
}
