package repotest

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suiteserve/suiteserve/internal/repo"
	"testing"
)

func TestRepo_Suite(t *testing.T) {
	r := Open(t)
	_, err := r.Suite("nonexistent")
	assert.True(t, errors.Is(err, repo.ErrNotFound), "want ErrNotFound")

	s := repo.Suite{
		Name: "test",
	}
	id, err := r.InsertSuite(s)
	require.Nil(t, err)
	s.Id = id

	got, err := r.Suite(id)
	require.Nil(t, err)
	assert.Equal(t, &s, got)
}
