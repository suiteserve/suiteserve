package repotest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suiteserve/suiteserve/internal/repo"
	"strconv"
	"testing"
)

type setQueryCall struct {
	id    string
	padLt int
	padGt int
	want  []repo.Change
}

var setQueryTests = []struct {
	suites        []repo.Suite
	setQueryCalls func(ids []string) []setQueryCall
}{
	{
		setQueryCalls: func(ids []string) []setQueryCall {
			return []setQueryCall{
				{
					want: []repo.Change{
						repo.SuiteAggUpdate{
							SuiteAgg: repo.SuiteAgg{},
						},
					},
				},
			}
		},
	},
	{
		setQueryCalls: func(ids []string) []setQueryCall {
			return []setQueryCall{
				{
					padLt: 2,
					padGt: 4,
					want: []repo.Change{
						repo.SuiteAggUpdate{
							SuiteAgg: repo.SuiteAgg{},
						},
					},
				},
			}
		},
	},
	{
		suites: []repo.Suite{{}},
		setQueryCalls: func(ids []string) []setQueryCall {
			return []setQueryCall{
				{
					id: ids[0],
					want: []repo.Change{
						repo.SuiteUpsert{
							Suite: repo.Suite{Entity: repo.Entity{Id: ids[0]}},
						},
						repo.SuiteAggUpdate{
							SuiteAgg: repo.SuiteAgg{
								VersionedEntity: repo.VersionedEntity{
									Version: 1,
								},
								TotalCount:   1,
								StartedCount: 0,
							},
						},
					},
				},
			}
		},
	},
	{
		suites: []repo.Suite{
			{StartedAt: 100},
			{StartedAt: 200},
			{StartedAt: 300},
			{StartedAt: 400},
			{
				Status:    repo.SuiteStatusStarted,
				StartedAt: 500,
			},
		},
		setQueryCalls: func(ids []string) []setQueryCall {
			return []setQueryCall{
				{
					id: ids[2],
					padGt: 1,
					padLt: 3,
					want: []repo.Change{
						repo.SuiteUpsert{
							Suite: repo.Suite{
								Entity: repo.Entity{Id: ids[3]},
								StartedAt: 400,
							},
						},
						repo.SuiteUpsert{
							Suite: repo.Suite{
								Entity: repo.Entity{Id: ids[2]},
								StartedAt: 300,
							},
						},
						repo.SuiteUpsert{
							Suite: repo.Suite{
								Entity: repo.Entity{Id: ids[1]},
								StartedAt: 200,
							},
						},
						repo.SuiteUpsert{
							Suite: repo.Suite{
								Entity: repo.Entity{Id: ids[0]},
								StartedAt: 100,
							},
						},
						repo.SuiteAggUpdate{
							SuiteAgg: repo.SuiteAgg{
								VersionedEntity: repo.VersionedEntity{
									Version: 5,
								},
								TotalCount:   5,
								StartedCount: 1,
							},
						},
					},
				},
				{
					id: ids[3],
					want: []repo.Change{
						repo.SuiteDelete{Id: ids[3]},
						repo.SuiteDelete{Id: ids[2]},
						repo.SuiteDelete{Id: ids[1]},
						repo.SuiteDelete{Id: ids[0]},
						repo.SuiteUpsert{
							Suite: repo.Suite{
								Entity: repo.Entity{Id: ids[3]},
								StartedAt: 400,
							},
						},
						repo.SuiteAggUpdate{
							SuiteAgg: repo.SuiteAgg{
								VersionedEntity: repo.VersionedEntity{
									Version: 5,
								},
								TotalCount:   5,
								StartedCount: 1,
							},
						},
					},
				},
				{
					padLt: 1,
					padGt: 1,
					want: []repo.Change{
						repo.SuiteDelete{Id: ids[3]},
						repo.SuiteUpsert{
							Suite: repo.Suite{
								Entity: repo.Entity{Id: ids[4]},
								Status: repo.SuiteStatusStarted,
								StartedAt: 500,
							},
						},
						repo.SuiteUpsert{
							Suite: repo.Suite{
								Entity: repo.Entity{Id: ids[3]},
								StartedAt: 400,
							},
						},
						repo.SuiteUpsert{
							Suite: repo.Suite{
								Entity: repo.Entity{Id: ids[2]},
								StartedAt: 300,
							},
						},
						repo.SuiteAggUpdate{
							SuiteAgg: repo.SuiteAgg{
								VersionedEntity: repo.VersionedEntity{
									Version: 5,
								},
								TotalCount:   5,
								StartedCount: 1,
							},
						},
					},
				},
			}
		},
	},
}

func TestWatchSuites_SetQuery(t *testing.T) {
	for i, test := range setQueryTests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			r := Open(t)

			var ids []string
			for _, s := range test.suites {
				id, err := r.InsertSuite(s)
				require.Nil(t, err)
				ids = append(ids, id)
			}

			w := r.WatchSuites()
			t.Cleanup(w.Close)
			for _, call := range test.setQueryCalls(ids) {
				require.Nil(t, w.SetQuery(call.id, call.padLt, call.padGt))
				assert.ElementsMatch(t, call.want, <-w.Changes())
			}
		})
	}
}
