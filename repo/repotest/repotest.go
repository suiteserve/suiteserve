package repotest

import (
	"github.com/suiteserve/suiteserve/repo"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func OpenBuntDb(t *testing.T) *repo.BuntDb {
	dir := newTempDir(t)
	db, err := repo.OpenBuntDb(filepath.Join(dir, "bunt.db"), &repo.FileRepo{
		Pattern: filepath.Join(dir, "*.attachment"),
	})
	if err != nil {
		t.Fatalf("open buntdb: %v\n", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close buntdb: %v\n", err)
		}
	})
	return db
}

func newTempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("create temp dir: %v\n", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("remove temp dir: %v\n", err)
		}
	})
	return dir
}
