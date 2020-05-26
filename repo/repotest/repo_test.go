package repotest

import (
	util "github.com/tmazeika/testpass/internal"
	"github.com/tmazeika/testpass/repo"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

const changeTimeout = 3 * time.Second

func TestRepo(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	util.RequireNil(t, err)
	defer func() {
		util.RequireNil(t, os.RemoveAll(dir))
	}()

	repoTests := func(t *testing.T, repos repo.Repos) {
		t.Run("Attachments", func(t *testing.T) {
			t.Run("Save_Find", countChangesTest(
				repos.Changes(),
				attachmentsSaveFind(repos.Attachments()),
				repo.ChangeOpInsert,
				repo.AttachmentColl,
				len(testAttachments)))
			t.Run("Find*_Delete*", countChangesTest(
				repos.Changes(),
				attachmentsFindDelete(repos.Attachments()),
				repo.ChangeOpUpdate,
				repo.AttachmentColl,
				3))
			t.Run("FindAll", attachmentsFindAll(repos.Attachments()))
		})
	}

	t.Run("BuntDB", func(t *testing.T) {
		repos, err := repo.OpenBuntRepos(filepath.Join(dir, "bunt.db"),
			filepath.Join(dir, "*.attachment"), repo.IncIntIdGenerator)
		util.RequireNil(t, err)
		defer func() {
			util.RequireNil(t, repos.Close())
		}()
		repoTests(t, repos)
	})
}

func countChangesTest(changeCh <-chan repo.Change, test func(t *testing.T), op repo.ChangeOp, coll repo.Collection, n int) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		var changes []repo.Change
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()

			timer := time.NewTimer(changeTimeout)
			defer timer.Stop()

			for len(changes) < n {
				select {
				case c, ok := <-changeCh:
					if !ok {
						return
					}
					changes = append(changes, c)

					if !timer.Stop() {
						<-timer.C
					}
					timer.Reset(changeTimeout)
				case <-timer.C:
					t.Fatalf("change timeout expired after %.1f seconds",
						changeTimeout.Seconds())
				}
			}
		}()

		test(t)
		wg.Wait()
		count := 0
		for _, c := range changes {
			if c.Op == op && c.Coll == coll {
				count++
			}
		}
		if count != n {
			t.Errorf("want %d '%s %s' changes, got %d", n, op, coll, count)
		}
	}
}
