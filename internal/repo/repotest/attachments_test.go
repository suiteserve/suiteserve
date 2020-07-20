package repotest

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suiteserve/suiteserve/internal/repo"
	"testing"
)

func TestRepo_Attachment(t *testing.T) {
	r := Open(t)
	_, err := r.Attachment("nonexistent")
	assert.True(t, errors.Is(err, repo.ErrNotFound), "want ErrNotFound")

	a := repo.Attachment{
		SoftDeleteEntity: repo.SoftDeleteEntity{
			Deleted:   true,
			DeletedAt: 1594999447324,
		},
		SuiteId:   "123",
		Filename:  "test.txt",
		Timestamp: 1594997447324,
	}
	id, err := r.InsertAttachment(a)
	require.Nil(t, err)
	a.Id = id

	got, err := r.Attachment(id)
	require.Nil(t, err)
	assert.Equal(t, &a, got)
}

func TestRepo_SuiteAttachments(t *testing.T) {
	r := Open(t)
	all, err := r.SuiteAttachments("123")
	require.Nil(t, err)
	assert.NotNil(t, all)
	assert.Empty(t, all)

	a1 := repo.Attachment{
		SuiteId:   "123",
		Filename:  "test.txt",
		Timestamp: 1594997447324,
	}
	id1, err := r.InsertAttachment(a1)
	require.Nil(t, err)
	a1.Id = id1

	a2 := repo.Attachment{
		SuiteId: "123",
	}
	id2, err := r.InsertAttachment(a2)
	require.Nil(t, err)
	a2.Id = id2

	_, err = r.InsertAttachment(repo.Attachment{
		CaseId: "123",
	})
	require.Nil(t, err)

	got, err := r.SuiteAttachments("123")
	require.Nil(t, err)
	assert.Equal(t, []*repo.Attachment{&a2, &a1}, got)
}

func TestRepo_CaseAttachments(t *testing.T) {
	r := Open(t)
	all, err := r.CaseAttachments("123")
	require.Nil(t, err)
	assert.NotNil(t, all)
	assert.Empty(t, all)

	a1 := repo.Attachment{
		CaseId:    "123",
		Filename:  "test.txt",
		Timestamp: 1594997447324,
	}
	id1, err := r.InsertAttachment(a1)
	require.Nil(t, err)
	a1.Id = id1

	a2 := repo.Attachment{
		CaseId: "123",
	}
	id2, err := r.InsertAttachment(a2)
	require.Nil(t, err)
	a2.Id = id2

	_, err = r.InsertAttachment(repo.Attachment{
		SuiteId: "123",
	})
	require.Nil(t, err)

	got, err := r.CaseAttachments("123")
	require.Nil(t, err)
	assert.Equal(t, []*repo.Attachment{&a2, &a1}, got)
}
