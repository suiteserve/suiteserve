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
	attachments []repo.Attachment
	setFilter   func(ids []string, r *pb.GetAttachmentsRequest)
	wantCount   int
	want        func(ids []string) []*pb.Attachment
}{
	{
		attachments: []repo.Attachment{{Filename: "test.txt"}},
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
		attachments: []repo.Attachment{
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
		attachments: []repo.Attachment{
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
		attachments: []repo.Attachment{
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
				id, err := r.InsertAttachment(a)
				require.Nil(t, err)
				ids = append(ids, id)
			}

			var in pb.GetAttachmentsRequest
			test.setFilter(ids, &in)
			got, err := client.GetAttachments(context.Background(), &in)
			require.Nil(t, err)
			if assert.Len(t, got.Attachments, test.wantCount) {
				assert.Equal(t, test.want(ids), got.Attachments)
			}
		})
	}
}
