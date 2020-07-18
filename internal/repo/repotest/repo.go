package repotest

import (
	"github.com/stretchr/testify/require"
	"github.com/suiteserve/suiteserve/internal/repo"
	"io/ioutil"
	"os"
	"testing"
)

func Open(t *testing.T) *repo.Repo {
	f, err := ioutil.TempFile("", "")
	require.Nil(t, err)
	filename := f.Name()
	t.Cleanup(func() {
		require.Nil(t, os.Remove(filename))
	})
	require.Nil(t, f.Close())
	r, err := repo.Open(filename)
	require.Nil(t, err)
	t.Cleanup(func() {
		require.Nil(t, r.Close())
	})
	return r
}
