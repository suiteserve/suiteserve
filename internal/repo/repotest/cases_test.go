package repotest

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suiteserve/suiteserve/internal/repo"
	"testing"
)

func TestRepo_Case(t *testing.T) {
	r := Open(t)
	_, err := r.Case("nonexistent")
	assert.True(t, errors.Is(err, repo.ErrNotFound), "want ErrNotFound")

	c := repo.Case{
		Name: "test",
	}
	id, err := r.InsertCase(c)
	require.Nil(t, err)
	c.Id = id

	got, err := r.Case(id)
	require.Nil(t, err)
	assert.Equal(t, &c, got)
}
