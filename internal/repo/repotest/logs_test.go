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

	l := repo.LogLine{
		Message: "Hello, world!",
	}
	id, err := r.InsertLogLine(l)
	require.Nil(t, err)
	l.Id = id

	got, err := r.LogLine(id)
	require.Nil(t, err)
	assert.Equal(t, &l, got)
}
