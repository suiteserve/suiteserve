package repotest

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/suiteserve/suiteserve/repo"
	"reflect"
	"testing"
)

func init() {
	spew.Config.Indent = "  "
	spew.Config.DisablePointerAddresses = true
	spew.Config.DisableCapacities = true
}

// order is [1, 2, 0] when sorted
var testSuites = []repo.UnsavedSuite{
	{},
	{
		SoftDeleteEntity: repo.SoftDeleteEntity{
			Deleted:   true,
			DeletedAt: 99999,
		},
		Name: "Lorem Ipsum",
		Tags: []string{"hello", "world"},
		EnvVars: []repo.SuiteEnvVar{
			{
				Key:   "key",
				Value: "val",
			},
			{
				Key:   "",
				Value: nil,
			},
		},
		PlannedCases: -5,
		Status:       "running",
		StartedAt:    1000000000,
	},
	{
		Tags:           []string{""},
		PlannedCases:   10000,
		StartedAt:      1,
		DisconnectedAt: -234,
	},
}

func TestBuntDb_InsertSuite_Suite(t *testing.T) {
	db := OpenBuntDb(t)

	for _, want := range testSuites {
		id, err := db.InsertSuite(context.Background(), &want)
		if err != nil {
			t.Fatalf("insert suite: %v", err)
		}
		want := &repo.Suite{
			SavedEntity: repo.SavedEntity{
				Id: id,
			},
			VersionedEntity: repo.VersionedEntity{
				Version: 0,
			},
			UnsavedSuite: want,
		}
		got, err := db.Suite(context.Background(), id)
		if err != nil {
			t.Fatalf("get suite: %v", err)
		}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want\n%vgot\n%v", spew.Sdump(want), spew.Sdump(got))
		}
	}
}

var testSuitePages = []struct {
	name string
	got  func(db *repo.BuntDb, ids []string) (*repo.SuitePage, error)
	want func(ids []string) *repo.SuitePage
}{
	{
		name: "limit 0",
		got: func(db *repo.BuntDb, _ []string) (*repo.SuitePage, error) {
			return db.SuitePage(context.Background(), "", 0)
		},
		want: func(ids []string) *repo.SuitePage {
			return &repo.SuitePage{
				NextId: &ids[2],
			}
		},
	},
	{
		name: "limit 5",
		got: func(db *repo.BuntDb, _ []string) (*repo.SuitePage, error) {
			return db.SuitePage(context.Background(), "", 5)
		},
		want: func(ids []string) *repo.SuitePage {
			return &repo.SuitePage{
				NextId: nil,
				Suites: []repo.Suite{
					{
						SavedEntity: repo.SavedEntity{
							Id: ids[2],
						},
						VersionedEntity: repo.VersionedEntity{
							Version: 0,
						},
						UnsavedSuite: testSuites[2],
					},
					{
						SavedEntity: repo.SavedEntity{
							Id: ids[0],
						},
						VersionedEntity: repo.VersionedEntity{
							Version: 0,
						},
						UnsavedSuite: testSuites[0],
					},
				},
			}
		},
	},
	{
		name: "from middle, limit 1",
		got: func(db *repo.BuntDb, ids []string) (*repo.SuitePage, error) {
			return db.SuitePage(context.Background(), ids[2], 1)
		},
		want: func(ids []string) *repo.SuitePage {
			return &repo.SuitePage{
				NextId: &ids[0],
				Suites: []repo.Suite{
					{
						SavedEntity: repo.SavedEntity{
							Id: ids[2],
						},
						VersionedEntity: repo.VersionedEntity{
							Version: 0,
						},
						UnsavedSuite: testSuites[2],
					},
				},
			}
		},
	},
	{
		name: "from deleted",
		got: func(db *repo.BuntDb, ids []string) (*repo.SuitePage, error) {
			return db.SuitePage(context.Background(), ids[1], 3)
		},
		want: func(ids []string) *repo.SuitePage {
			return &repo.SuitePage{
				NextId: nil,
				Suites: []repo.Suite{
					{
						SavedEntity: repo.SavedEntity{
							Id: ids[2],
						},
						VersionedEntity: repo.VersionedEntity{
							Version: 0,
						},
						UnsavedSuite: testSuites[2],
					},
					{
						SavedEntity: repo.SavedEntity{
							Id: ids[0],
						},
						VersionedEntity: repo.VersionedEntity{
							Version: 0,
						},
						UnsavedSuite: testSuites[0],
					},
				},
			}
		},
	},
}

func TestBuntDb_SuitePage(t *testing.T) {
	db := OpenBuntDb(t)

	ids := make([]string, len(testSuites))
	for i, s := range testSuites {
		id, err := db.InsertSuite(context.Background(), &s)
		if err != nil {
			t.Fatalf("insert suite: %v", err)
		}
		ids[i] = id
	}

	for _, test := range testSuitePages {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := test.got(db, ids)
			if err != nil {
				t.Fatalf("get suite page: %v", err)
			}
			want := test.want(ids)
			want.Aggs = repo.SuiteAggs{
				VersionedEntity: repo.VersionedEntity{
					Version: 3,
				},
				Running:  1,
				Finished: 2,
			}
			if !reflect.DeepEqual(want, got) {
				t.Errorf("want\n%vgot\n%v", spew.Sdump(want), spew.Sdump(got))
			}
		})
	}
}
