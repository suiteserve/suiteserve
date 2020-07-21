package repotest

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suiteserve/suiteserve/internal/repo"
	"strconv"
	"testing"
)

func TestRepo_Suite(t *testing.T) {
	r := Open(t)
	_, err := r.Suite("nonexistent")
	assert.True(t, errors.Is(err, repo.ErrNotFound), "want ErrNotFound")

	s := repo.Suite{
		Name: "test",
	}
	id, err := r.InsertSuite(s)
	require.Nil(t, err)
	s.Id = id

	got, err := r.Suite(id)
	require.Nil(t, err)
	assert.Equal(t, &s, got)
}

var suitePageTests = []struct {
	suites    []*repo.Suite
	fromId    func(ids []string) string
	limit     int
	wantCount int
	want      func(ids []string) *repo.SuitePage
}{
	{
		suites: []*repo.Suite{},
		fromId: func(ids []string) string {
			return "abc"
		},
		limit:     0,
		wantCount: 0,
	},
	{
		suites: []*repo.Suite{},
		fromId: func(ids []string) string {
			return "abc"
		},
		limit:     3,
		wantCount: 0,
	},
	{
		suites: []*repo.Suite{{Name: "test1"}},
		fromId: func(ids []string) string {
			return ids[0]
		},
		limit:     0,
		wantCount: 1,
		want: func(ids []string) *repo.SuitePage {
			return &repo.SuitePage{
				SuiteAgg: repo.SuiteAgg{
					VersionedEntity: repo.VersionedEntity{Version: 1},
					TotalCount:      1,
					StartedCount:    0,
				},
				HasMore: false,
				Suites: []*repo.Suite{{
					Entity: repo.Entity{Id: ids[0]},
					Name:   "test1",
				}},
			}
		},
	},
	{
		suites: []*repo.Suite{
			{StartedAt: 400},
			{StartedAt: 200},
			{
				Status:    repo.SuiteStatusStarted,
				StartedAt: 300,
			},
			{StartedAt: 100},
		},
		fromId: func(ids []string) string {
			return ""
		},
		limit:     3,
		wantCount: 3,
		want: func(ids []string) *repo.SuitePage {
			return &repo.SuitePage{
				SuiteAgg: repo.SuiteAgg{
					VersionedEntity: repo.VersionedEntity{Version: 4},
					TotalCount:      4,
					StartedCount:    1,
				},
				HasMore: true,
				Suites: []*repo.Suite{
					{
						Entity:    repo.Entity{Id: ids[0]},
						StartedAt: 400,
					},
					{
						Entity:    repo.Entity{Id: ids[2]},
						Status:    repo.SuiteStatusStarted,
						StartedAt: 300,
					},
					{
						Entity:    repo.Entity{Id: ids[1]},
						StartedAt: 200,
					},
				},
			}
		},
	},
	{
		suites: []*repo.Suite{
			{StartedAt: 400},
			{StartedAt: 200},
			{
				Status:    repo.SuiteStatusStarted,
				StartedAt: 300,
			},
			{StartedAt: 100},
		},
		fromId: func(ids []string) string {
			return ""
		},
		limit:     4,
		wantCount: 4,
		want: func(ids []string) *repo.SuitePage {
			return &repo.SuitePage{
				SuiteAgg: repo.SuiteAgg{
					VersionedEntity: repo.VersionedEntity{Version: 4},
					TotalCount:      4,
					StartedCount:    1,
				},
				HasMore: false,
				Suites: []*repo.Suite{
					{
						Entity:    repo.Entity{Id: ids[0]},
						StartedAt: 400,
					},
					{
						Entity:    repo.Entity{Id: ids[2]},
						Status:    repo.SuiteStatusStarted,
						StartedAt: 300,
					},
					{
						Entity:    repo.Entity{Id: ids[1]},
						StartedAt: 200,
					},
					{
						Entity:    repo.Entity{Id: ids[3]},
						StartedAt: 100,
					},
				},
			}
		},
	},
}

func TestRepo_SuitePage(t *testing.T) {
	for i, test := range suitePageTests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			r := Open(t)

			var ids []string
			for _, s := range test.suites {
				id, err := r.InsertSuite(*s)
				require.Nil(t, err)
				ids = append(ids, id)
			}

			got, err := r.SuitePage(test.fromId(ids), test.limit)
			if test.wantCount == 0 {
				assert.True(t, errors.Is(err, repo.ErrNotFound), "want ErrNotFound")
				return
			}
			require.Nil(t, err)
			if assert.Len(t, got.Suites, test.wantCount) {
				assert.Equal(t, test.want(ids), got)
			}
		})
	}
}
