package repotest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suiteserve/suiteserve/internal/repo"
	"testing"
)

func TestRepo_Case(t *testing.T) {
	r := Open(t)
	_, err := r.Case("nonexistent")
	assert.True(t, isNotFound(err), "want not found")

	want := repo.Case{
		Name: "test",
	}
	id, err := r.InsertCase(want)
	require.Nil(t, err)
	want.Id = id

	got, err := r.Case(id)
	require.Nil(t, err)
	assert.Equal(t, want, got)
}
