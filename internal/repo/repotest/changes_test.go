package repotest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suiteserve/suiteserve/internal/repo"
	"strconv"
	"testing"
)

type setQueryInputWant struct {
	id    string
	padLt int
	padGt int
	want  []repo.Change
}

var setQueryTests = []struct {
	suites []repo.Suite
	iw     func(ids []string) setQueryInputWant
}{
	{
		iw: func(ids []string) setQueryInputWant {
			return setQueryInputWant{
				want: []repo.Change{
					repo.SuiteAggUpsert{
						SuiteAgg: repo.SuiteAgg{},
					},
				},
			}
		},
	},
	{
		iw: func(ids []string) setQueryInputWant {
			return setQueryInputWant{
				padLt: 2,
				padGt: 4,
				want: []repo.Change{
					repo.SuiteAggUpsert{
						SuiteAgg: repo.SuiteAgg{},
					},
				},
			}
		},
	},
	{
		suites: []repo.Suite{{}},
		iw: func(ids []string) setQueryInputWant {
			return setQueryInputWant{
				id: ids[0],
				want: []repo.Change{
					repo.SuiteUpsert{
						Suite: repo.Suite{Entity: repo.Entity{Id: ids[0]}},
					},
					repo.SuiteAggUpsert{
						SuiteAgg: repo.SuiteAgg{
							VersionedEntity: repo.VersionedEntity{
								Version: 1,
							},
							TotalCount:   1,
							StartedCount: 0,
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
		iw: func(ids []string) setQueryInputWant {
			return setQueryInputWant{
				id:    ids[2],
				padGt: 1,
				padLt: 3,
				want: []repo.Change{
					repo.SuiteUpsert{
						Suite: repo.Suite{
							Entity:    repo.Entity{Id: ids[3]},
							StartedAt: 400,
						},
					},
					repo.SuiteUpsert{
						Suite: repo.Suite{
							Entity:    repo.Entity{Id: ids[2]},
							StartedAt: 300,
						},
					},
					repo.SuiteUpsert{
						Suite: repo.Suite{
							Entity:    repo.Entity{Id: ids[1]},
							StartedAt: 200,
						},
					},
					repo.SuiteUpsert{
						Suite: repo.Suite{
							Entity:    repo.Entity{Id: ids[0]},
							StartedAt: 100,
						},
					},
					repo.SuiteAggUpsert{
						SuiteAgg: repo.SuiteAgg{
							VersionedEntity: repo.VersionedEntity{
								Version: 5,
							},
							TotalCount:   5,
							StartedCount: 1,
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

			iw := test.iw(ids)
			w, err := r.WatchSuites(iw.id, iw.padLt, iw.padGt)
			require.Nil(t, err)
			t.Cleanup(w.Close)
			assert.ElementsMatch(t, iw.want, <-w.Changes())
		})
	}
}
