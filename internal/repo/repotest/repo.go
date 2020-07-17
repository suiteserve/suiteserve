package repotest

import (
	"github.com/suiteserve/suiteserve/internal/repo"
	"io/ioutil"
	"os"
	"testing"
)

func Open(t *testing.T) *repo.Repo {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	filename := f.Name()
	t.Cleanup(func() {
		if err := os.Remove(filename); err != nil {
			t.Fatal(err)
		}
	})
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	r, err := repo.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
	})
	return r
}
