package repotest

import (
	"errors"
	"github.com/stretchr/testify/require"
	"github.com/suiteserve/suiteserve/internal/repo"
	"testing"
)

func isNotFound(err error) bool {
	var foundErr interface {
		Found() bool
	}
	return errors.As(err, &foundErr) && !foundErr.Found()
}

func Open(t *testing.T) *repo.Repo {
	r, err := repo.Open(":memory:")
	require.Nil(t, err)
	t.Cleanup(func() {
		require.Nil(t, r.Close())
	})
	return r
}
