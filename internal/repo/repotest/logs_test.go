package repotest

import (
	"errors"
	"github.com/suiteserve/suiteserve/internal/repo"
	"reflect"
	"testing"
)

func TestRepo_LogLine(t *testing.T) {
	r := Open(t)
	_, err := r.LogLine("nonexistent")
	if !errors.Is(err, repo.ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}

	l := repo.LogLine{
		Message: "Hello, world!",
	}
	id, err := r.InsertLogLine(l)
	if err != nil {
		t.Fatalf("insert log line: %v", err)
	}

	got, err := r.LogLine(id)
	if err != nil {
		t.Fatalf("get log line: %v", err)
	}
	l.Id = id
	want := &l
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
