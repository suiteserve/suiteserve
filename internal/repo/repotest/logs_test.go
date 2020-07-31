package repotest

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suiteserve/suiteserve/internal/repo"
	"testing"
)

func TestRepo_LogLine(t *testing.T) {
	r := Open(t)
	_, err := r.LogLine("nonexistent")
	assert.True(t, errors.Is(err, repo.ErrNotFound), "want ErrNotFound")

	want := repo.LogLine{
		Message: "Hello, world!",
	}
	id, err := r.InsertLogLine(want)
	require.Nil(t, err)
	want.Id = id

	got, err := r.LogLine(id)
	require.Nil(t, err)
	assert.Equal(t, want, got)
}
