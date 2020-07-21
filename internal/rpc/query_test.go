package rpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/suiteserve/protocol/go/protocol"
	"github.com/suiteserve/suiteserve/internal/repo"
	"github.com/suiteserve/suiteserve/internal/repo/repotest"
	"strconv"
	"testing"
)

var getAttachmentsTests = []struct {
	attachments []*repo.Attachment
	setFilter   func(ids []string, r *pb.GetAttachmentsRequest)
	wantCount   int
	want        func(ids []string) []*pb.Attachment
}{
	{
		attachments: []*repo.Attachment{{Filename: "test.txt"}},
		setFilter: func(ids []string, r *pb.GetAttachmentsRequest) {
			r.Filter = &pb.GetAttachmentsRequest_Id{Id: ids[0]}
		},
		wantCount: 1,
		want: func(ids []string) []*pb.Attachment {
			return []*pb.Attachment{{
				Id:        ids[0],
				Filename:  "test.txt",
				DeletedAt: millisToPb(0),
				Timestamp: millisToPb(0),
			}}
		},
	},
	{
		attachments: []*repo.Attachment{
			{Filename: "test.txt"},
			{Url: "https://example.com/test.txt"},
		},
		setFilter: func(ids []string, r *pb.GetAttachmentsRequest) {
			r.Filter = &pb.GetAttachmentsRequest_Id{Id: ids[1]}
		},
		wantCount: 1,
		want: func(ids []string) []*pb.Attachment {
			return []*pb.Attachment{{
				Id:        ids[1],
				Url:       "https://example.com/test.txt",
				DeletedAt: millisToPb(0),
				Timestamp: millisToPb(0),
			}}
		},
	},
	{
		attachments: []*repo.Attachment{
			{
				SuiteId:   "123",
				Timestamp: 100,
			},
			{
				CaseId:    "123",
				Timestamp: 200,
			},
			{
				SuiteId:   "123",
				Timestamp: 300,
			},
		},
		setFilter: func(_ []string, r *pb.GetAttachmentsRequest) {
			r.Filter = &pb.GetAttachmentsRequest_SuiteId{SuiteId: "123"}
		},
		wantCount: 2,
		want: func(ids []string) []*pb.Attachment {
			return []*pb.Attachment{
				{
					Id:        ids[2],
					Owner:     &pb.Attachment_SuiteId{SuiteId: "123"},
					DeletedAt: millisToPb(0),
					Timestamp: millisToPb(300),
				},
				{
					Id:        ids[0],
					Owner:     &pb.Attachment_SuiteId{SuiteId: "123"},
					DeletedAt: millisToPb(0),
					Timestamp: millisToPb(100),
				},
			}
		},
	},
	{
		attachments: []*repo.Attachment{
			{
				CaseId:    "123",
				Timestamp: 100,
			},
			{
				SuiteId:   "123",
				Timestamp: 200,
			},
			{
				CaseId:    "123",
				Timestamp: 300,
			},
		},
		setFilter: func(_ []string, r *pb.GetAttachmentsRequest) {
			r.Filter = &pb.GetAttachmentsRequest_CaseId{CaseId: "123"}
		},
		wantCount: 2,
		want: func(ids []string) []*pb.Attachment {
			return []*pb.Attachment{
				{
					Id:        ids[2],
					Owner:     &pb.Attachment_CaseId{CaseId: "123"},
					DeletedAt: millisToPb(0),
					Timestamp: millisToPb(300),
				},
				{
					Id:        ids[0],
					Owner:     &pb.Attachment_CaseId{CaseId: "123"},
					DeletedAt: millisToPb(0),
					Timestamp: millisToPb(100),
				},
			}
		},
	},
}

func TestQuery_GetAttachments(t *testing.T) {
	for i, test := range getAttachmentsTests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			r := repotest.Open(t)
			client := pb.NewQueryServiceClient(newClientConn(t, r))

			var ids []string
			for _, a := range test.attachments {
				id, err := r.InsertAttachment(*a)
				require.Nil(t, err)
				ids = append(ids, id)
			}

			in := &pb.GetAttachmentsRequest{}
			test.setFilter(ids, in)
			got, err := client.GetAttachments(context.Background(), in)
			require.Nil(t, err)
			if assert.Len(t, got.Attachments, test.wantCount) {
				want := test.want(ids)
				assert.Equal(t, want, got.Attachments)
			}
		})
	}
}

type insertOutput struct {
	removed []watchHandle
	ok      bool
	window  []watchHandle
}

var insertTests = []struct {
	w    *watchWindow
	in   watchHandle
	want insertOutput
}{
	{
		w:  &watchWindow{},
		in: watchHandle{id: "1", val: 1},
		want: insertOutput{
			ok:     true,
			window: []watchHandle{{id: "1", val: 1}},
		},
	},
	{
		w: &watchWindow{
			minSize: 2,
		},
		in: watchHandle{id: "1", val: 1},
		want: insertOutput{
			ok:     true,
			window: []watchHandle{{id: "1", val: 1}},
		},
	},
	{
		w: &watchWindow{
			required: &watchHandle{id: "1", val: 1},
			minSize:  3,
		},
		in: watchHandle{id: "1", val: 1},
		want: insertOutput{
			ok:     true,
			window: []watchHandle{{id: "1", val: 1}},
		},
	},
	{
		w: &watchWindow{
			required: &watchHandle{id: "22", val: 2},
			minSize:  2,
			window: []watchHandle{
				{id: "10", val: 1},
				{id: "20", val: 2},
				{id: "22", val: 2},
				{id: "30", val: 3},
			},
		},
		in: watchHandle{id: "21", val: 2},
		want: insertOutput{
			removed: []watchHandle{{id: "10", val: 1}},
			ok:      true,
			window: []watchHandle{
				{id: "21", val: 2},
				{id: "20", val: 2},
				{id: "22", val: 2},
				{id: "30", val: 3},
			},
		},
	},
	{
		w: &watchWindow{
			required: &watchHandle{id: "22", val: 2},
			minSize:  3,
			window: []watchHandle{
				{id: "20", val: 2},
				{id: "22", val: 2},
				{id: "30", val: 3},
			},
		},
		in: watchHandle{id: "21", val: 2},
		want: insertOutput{
			ok: true,
			window: []watchHandle{
				{id: "21", val: 2},
				{id: "20", val: 2},
				{id: "22", val: 2},
				{id: "30", val: 3},
			},
		},
	},
	{
		w: &watchWindow{
			required: &watchHandle{id: "22", val: 2},
			minSize:  3,
			window: []watchHandle{
				{id: "20", val: 2},
				{id: "22", val: 2},
				{id: "30", val: 3},
			},
		},
		in: watchHandle{id: "10", val: 1},
		want: insertOutput{
			ok: false,
			window: []watchHandle{
				{id: "20", val: 2},
				{id: "22", val: 2},
				{id: "30", val: 3},
			},
		},
	},
}

func TestWatchWindow_Insert(t *testing.T) {
	for i, test := range insertTests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			removed, ok := test.w.insert(test.in)
			assert.ElementsMatch(t, test.want.removed, removed)
			assert.Equal(t, test.want.ok, ok)
			assert.Equal(t, test.want.window, test.w.window)
		})
	}
}

var shrinkTests = []struct {
	required int
	minSize  int
	window   []int
	want     int
}{
	{
		minSize: 3,
		want:    0,
	},
	{
		minSize: 2,
		window:  []int{},
		want:    0,
	},
	{
		minSize: 0,
		window:  []int{1, 2, 3},
		want:    0,
	},
	{
		minSize: -1,
		window:  []int{2, 2, 9, 10},
		want:    0,
	},
	{
		minSize: 2,
		window:  []int{1, 2, 3, 3, 4},
		want:    2,
	},
	{
		required: 2,
		minSize:  2,
		window:   []int{1, 2, 3, 3, 4},
		want:     1,
	},
	{
		required: 3,
		minSize:  4,
		window:   []int{1, 2, 3, 3, 4},
		want:     1,
	},
	{
		minSize: 1,
		window:  []int{1, 1, 1, 1},
		want:    0,
	},
	{
		minSize: 2,
		window:  []int{1, 1, 2, 3, 3, 4},
		want:    3,
	},
	{
		required: 1,
		minSize:  2,
		window:   []int{1, 1, 2, 3, 3, 4},
		want:     0,
	},
	{
		minSize: 1,
		window:  []int{1, 1, 2, 3, 3, 4},
		want:    5,
	},
	{
		required: 4,
		minSize:  1,
		window:   []int{1, 1, 2, 3, 3, 4},
		want:     5,
	},
	{
		required: 2,
		minSize:  5,
		window:   []int{1, 1, 2, 3, 3, 4},
		want:     0,
	},
	{
		required: 1,
		minSize:  2,
		window:   []int{1, 1, 1},
		want:     0,
	},
}

func TestWatchWindow_Shrink(t *testing.T) {
	for i, test := range shrinkTests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			var w watchWindow
			if test.required > 0 {
				w.required = &watchHandle{val: int64(test.required)}
			}
			w.minSize = test.minSize
			for _, i := range test.window {
				w.window = append(w.window, watchHandle{val: int64(i)})
			}
			assert.Equal(t, test.want, w.shrink())
		})
	}
}
