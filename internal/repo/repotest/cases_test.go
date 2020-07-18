package repotest

import (
	"errors"
	"github.com/suiteserve/suiteserve/internal/repo"
	"reflect"
	"testing"
)

func TestRepo_Case(t *testing.T) {
	r := Open(t)
	_, err := r.Case("nonexistent")
	if !errors.Is(err, repo.ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}

	c := repo.Case{
		Name: "test",
	}
	id, err := r.InsertCase(c)
	if err != nil {
		t.Fatalf("insert case: %v", err)
	}

	got, err := r.Case(id)
	if err != nil {
		t.Fatalf("get case: %v", err)
	}
	c.Id = id
	want := &c
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
