package repotest

import (
	"github.com/stretchr/testify/require"
	"github.com/suiteserve/suiteserve/internal/repo"
	"testing"
)

func Open(t *testing.T) *repo.Repo {
	r, err := repo.Open(":memory:")
	require.Nil(t, err)
	t.Cleanup(func() {
		require.Nil(t, r.Close())
	})
	return r
}
