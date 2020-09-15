package repotest

// import (
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"github.com/suiteserve/suiteserve/internal/repo"
// 	"testing"
// )
//
// func TestRepo_Attachment(t *testing.T) {
// 	r := Open(t)
// 	_, err := r.Attachment("nonexistent")
// 	assert.True(t, isNotFound(err), "want not found")
//
// 	want := insertAttachment(t, r, repo.Attachment{
// 		SoftDeleteEntity: repo.SoftDeleteEntity{
// 			Deleted:   true,
// 			DeletedAt: 1594999447324,
// 		},
// 		SuiteId:   "123",
// 		Filename:  "test.txt",
// 		Timestamp: 1594997447324,
// 	})
//
// 	got, err := r.Attachment(want.Id)
// 	require.Nil(t, err)
// 	assert.Equal(t, want, got)
// }
//
// func TestRepo_SuiteAttachments(t *testing.T) {
// 	r := Open(t)
// 	all, err := r.SuiteAttachments("123")
// 	require.Nil(t, err)
// 	assert.NotNil(t, all)
// 	assert.Empty(t, all)
//
// 	want1 := insertAttachment(t, r, repo.Attachment{
// 		SuiteId:   "123",
// 		Filename:  "test.txt",
// 		Timestamp: 1594997447324,
// 	})
// 	want2 := insertAttachment(t, r, repo.Attachment{
// 		SuiteId: "123",
// 	})
// 	_ = insertAttachment(t, r, repo.Attachment{
// 		CaseId: "123",
// 	})
//
// 	got, err := r.SuiteAttachments("123")
// 	require.Nil(t, err)
// 	assert.Equal(t, []repo.Attachment{want2, want1}, got)
// }
//
// func TestRepo_CaseAttachments(t *testing.T) {
// 	r := Open(t)
// 	all, err := r.CaseAttachments("123")
// 	require.Nil(t, err)
// 	assert.NotNil(t, all)
// 	assert.Empty(t, all)
//
// 	want1 := insertAttachment(t, r, repo.Attachment{
// 		CaseId:    "123",
// 		Filename:  "test.txt",
// 		Timestamp: 1594997447324,
// 	})
// 	want2 := insertAttachment(t, r, repo.Attachment{
// 		CaseId: "123",
// 	})
// 	_ = insertAttachment(t, r, repo.Attachment{
// 		SuiteId: "123",
// 	})
//
// 	got, err := r.CaseAttachments("123")
// 	require.Nil(t, err)
// 	assert.Equal(t, []repo.Attachment{want2, want1}, got)
// }
//
// func insertAttachment(t *testing.T, r *repo.Repo, a repo.Attachment) repo.Attachment {
// 	t.Helper()
// 	id, err := r.InsertAttachment(a)
// 	require.Nil(t, err)
// 	a.Id = id
// 	return a
// }
