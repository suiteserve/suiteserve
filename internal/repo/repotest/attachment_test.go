package repotest

import (
	"errors"
	"github.com/suiteserve/suiteserve/internal/repo"
	"reflect"
	"testing"
	"time"
)

func TestRepo_Attachment(t *testing.T) {
	r := Open(t)
	_, err := r.Attachment("nonexistent")
	if !errors.Is(err, repo.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}

	a := repo.Attachment{
		SoftDeleteEntity: repo.SoftDeleteEntity{
			Deleted:   true,
			DeletedAt: time.Unix(1594999447, 324*1e6),
		},
		SuiteId:   "123",
		Filename:  "test.txt",
		Timestamp: time.Unix(1594997447, 324*1e6),
	}
	id, err := r.InsertAttachment(a)
	if err != nil {
		t.Fatalf("insert attachment: %v", err)
	}

	a.Id = id
	want := &a
	got, err := r.Attachment(id)
	if err != nil {
		t.Fatalf("get attachment: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want %+v, got %+v", want, got)
	}
}

func TestRepo_SuiteAttachments(t *testing.T) {
	r := Open(t)
	all, err := r.SuiteAttachments("123")
	if err != nil {
		t.Fatalf("get suite attachments: %v", err)
	}
	if len(all) > 0 {
		t.Errorf("want empty, got %v", all)
	}

	a1 := repo.Attachment{
		SuiteId:   "123",
		Filename:  "test.txt",
		Timestamp: time.Unix(1594997447, 324*1e6),
	}
	id1, err := r.InsertAttachment(a1)
	if err != nil {
		t.Fatalf("insert attachment: %v", err)
	}

	a2 := repo.Attachment{
		SuiteId: "123",
	}
	id2, err := r.InsertAttachment(a2)
	if err != nil {
		t.Fatalf("insert attachment: %v", err)
	}

	_, err = r.InsertAttachment(repo.Attachment{
		CaseId: "123",
	})
	if err != nil {
		t.Fatalf("insert attachment: %v", err)
	}

	a1.Id, a2.Id = id1, id2
	want := []*repo.Attachment{&a2, &a1}
	got, err := r.SuiteAttachments("123")
	if err != nil {
		t.Fatalf("get suite attachments: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want %+v, got %+v", want, got)
	}
}

func TestRepo_CaseAttachments(t *testing.T) {
	r := Open(t)
	all, err := r.CaseAttachments("123")
	if err != nil {
		t.Fatalf("get case attachments: %v", err)
	}
	if len(all) > 0 {
		t.Errorf("want empty, got %v", all)
	}

	a1 := repo.Attachment{
		CaseId:    "123",
		Filename:  "test.txt",
		Timestamp: time.Unix(1594997447, 324*1e6),
	}
	id1, err := r.InsertAttachment(a1)
	if err != nil {
		t.Fatalf("insert attachment: %v", err)
	}

	a2 := repo.Attachment{
		CaseId: "123",
	}
	id2, err := r.InsertAttachment(a2)
	if err != nil {
		t.Fatalf("insert attachment: %v", err)
	}

	_, err = r.InsertAttachment(repo.Attachment{
		SuiteId: "123",
	})
	if err != nil {
		t.Fatalf("insert attachment: %v", err)
	}

	a1.Id, a2.Id = id1, id2
	want := []*repo.Attachment{&a2, &a1}
	got, err := r.CaseAttachments("123")
	if err != nil {
		t.Fatalf("get case attachments: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want %+v, got %+v", want, got)
	}
}
