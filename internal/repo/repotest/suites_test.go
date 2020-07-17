package repotest

import (
	"errors"
	"github.com/suiteserve/suiteserve/internal/repo"
	"reflect"
	"testing"
)

func TestRepo_Suite(t *testing.T) {
	r := Open(t)
	_, err := r.Suite("nonexistent")
	if !errors.Is(err, repo.ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}

	s := repo.Suite{
		Name: "test",
	}
	id, err := r.InsertSuite(s)
	if err != nil {
		t.Fatalf("insert suite: %v", err)
	}

	got, err := r.Suite(id)
	if err != nil {
		t.Fatalf("get suite: %v", err)
	}
	s.Id = id
	want := &s
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
