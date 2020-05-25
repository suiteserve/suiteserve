package repo

import (
	"context"
	util "github.com/tmazeika/testpass/internal"
	"strconv"
	"testing"
)

var testAttachments = []Attachment{
	{},
	{
		SoftDeleteEntity: SoftDeleteEntity{
			Entity:    Entity{Id: "235\"."},
			Deleted:   false,
			DeletedAt: -5,
		},
		Filename:    "test.txt&mdash;",
		Size:        8,
		ContentType: "\"text/plain\"",
	},
	{
		Filename:    "",
		Size:        -5,
		ContentType: "tett/plain",
	},
	{
		SoftDeleteEntity: SoftDeleteEntity{
			Deleted:   true,
			DeletedAt: 0,
		},
		Filename:    "Hello.wor\n\tld",
		Size:        0,
		ContentType: "text/plai!~1`n; charset=UTF-8",
	},
	{
		SoftDeleteEntity: SoftDeleteEntity{
			Entity:    Entity{Id: ""},
			Deleted:   true,
			DeletedAt: 234562466,
		},
		Filename:    "%^  Hello.world",
		Size:        99328578,
		ContentType: "",
	},
}

func attachmentsSaveFind(repo AttachmentRepo) func(*testing.T) {
	return func(t *testing.T) {
		for i, expected := range testAttachments {
			id, err := repo.Save(context.Background(), expected)
			util.RequireNil(t, err)
			if id != strconv.Itoa(i+1) {
				t.Errorf("got %q, want %q", id, i+1)
			}

			actual, err := repo.Find(context.Background(), id)
			util.RequireNil(t, err)
			if actual.Id != id {
				t.Errorf("got %q, want %q", actual.Id, id)
			}
			assertAttachmentEquals(t, actual, &expected)
		}
	}
}

func attachmentsFindDelete(repo AttachmentRepo) func(*testing.T) {
	return func(t *testing.T) {
		attachments, err := repo.FindAll(context.Background(), true)
		util.RequireNil(t, err)

		if len(attachments) != len(testAttachments) {
			t.Errorf("got %d attachments, want %d",
				len(attachments), len(testAttachments))
		}
		for i, expected := range testAttachments {
			actual := attachments[i]
			assertAttachmentEquals(t, &actual, &expected)
		}

		const deletedAt = 100
		util.RequireNil(t,
			repo.Delete(context.Background(), attachments[1].Id, deletedAt))
		attachments, err = repo.FindAll(context.Background(), true)
		util.RequireNil(t, err)

		if len(attachments) != len(testAttachments) {
			t.Errorf("got %d attachments, want %d",
				len(attachments), len(testAttachments))
		}
		for i, expected := range testAttachments {
			if i == 1 {
				expected.Deleted = true
				expected.DeletedAt = deletedAt
			}
			actual := attachments[i]
			assertAttachmentEquals(t, &actual, &expected)
		}

		const deletedAt2 = 200
		util.RequireNil(t,
			repo.DeleteAll(context.Background(), deletedAt2))
		attachments, err = repo.FindAll(context.Background(), true)
		util.RequireNil(t, err)
		if len(attachments) != len(testAttachments) {
			t.Errorf("got %d attachments, want %d",
				len(attachments), len(testAttachments))
		}
		for i, expected := range testAttachments {
			if i == 1 {
				expected.DeletedAt = deletedAt
			} else if !expected.Deleted {
				expected.DeletedAt = deletedAt2
			}
			expected.Deleted = true

			actual := attachments[i]
			assertAttachmentEquals(t, &actual, &expected)

			actual2, err := repo.Find(context.Background(), actual.Id)
			util.RequireNil(t, err)
			assertAttachmentEquals(t, actual2, &expected)
		}
	}
}

func attachmentsFindAll(repo AttachmentRepo) func(*testing.T) {
	return func(t *testing.T) {
		attachments, err := repo.FindAll(context.Background(), false)
		util.RequireNil(t, err)

		if len(attachments) != 0 {
			t.Errorf("got %d attachments, want %d",
				len(attachments), 0)
		}
	}
}

func assertAttachmentEquals(t *testing.T, actual *Attachment, expected *Attachment) {
	t.Helper()
	if actual.Filename != expected.Filename {
		t.Errorf("got %q, want %q", actual.Filename, expected.Filename)
	}
	if actual.Size != expected.Size {
		t.Errorf("got %d, want %d", actual.Size, expected.Size)
	}
	if actual.ContentType != expected.ContentType {
		t.Errorf("got %q, want %q", actual.ContentType, expected.ContentType)
	}
	if actual.Deleted != expected.Deleted {
		t.Errorf("got %t, want %t", actual.Deleted, expected.Deleted)
	}
	if actual.DeletedAt != expected.DeletedAt {
		t.Errorf("got %d, want %d", actual.DeletedAt, expected.DeletedAt)
	}
}
