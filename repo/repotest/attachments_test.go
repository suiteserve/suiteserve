package repotest

import (
	"context"
	util "github.com/tmazeika/testpass/internal"
	"github.com/tmazeika/testpass/repo"
	"strconv"
	"strings"
	"testing"
)

var testAttachments = []repo.UnsavedAttachmentInfo{
	{},
	{
		SoftDeleteEntity: repo.SoftDeleteEntity{
			Deleted:   false,
			DeletedAt: -5,
		},
		Filename:    "test.txt&mdash;",
		ContentType: "\"text/plain\"",
	},
	{
		Filename:    "",
		ContentType: "tett/plain",
	},
	{
		SoftDeleteEntity: repo.SoftDeleteEntity{
			Deleted:   true,
			DeletedAt: 0,
		},
		Filename:    "Hello.wor\n\tld",
		ContentType: "text/plai!~1`n; charset=UTF-8",
	},
	{
		SoftDeleteEntity: repo.SoftDeleteEntity{
			Deleted:   true,
			DeletedAt: 234562466,
		},
		Filename:    "%^  Hello.world",
		ContentType: "",
	},
}

func attachmentsSaveFind(repo repo.AttachmentRepo) func(*testing.T) {
	return func(t *testing.T) {
		for i, expected := range testAttachments {
			file, err := repo.Save(context.Background(), expected,
				strings.NewReader("Hello, world!"))
			util.RequireNil(t, err)
			id := file.Info().Id
			if id != strconv.Itoa(i+1) {
				t.Errorf("got %q, want %q", id, i+1)
			}
			if file.Info().Size != 13 {
				t.Errorf("got size %d, want %d", file.Info().Size, 13)
			}
			assertAttachmentEquals(t, file.Info(), &expected)

			actual, err := repo.Find(context.Background(), id)
			util.RequireNil(t, err)
			if actual.Info().Id != id {
				t.Errorf("got %q, want %q", actual.Info().Id, id)
			}
			if file.Info().Size != 13 {
				t.Errorf("got size %d, want %d", file.Info().Size, 13)
			}
			assertAttachmentEquals(t, actual.Info(), &expected)
		}
	}
}

func attachmentsFindDelete(repo repo.AttachmentRepo) func(*testing.T) {
	return func(t *testing.T) {
		attachments, err := repo.FindAll(context.Background(), true)
		util.RequireNil(t, err)

		if len(attachments) != len(testAttachments) {
			t.Errorf("got %d attachments, want %d",
				len(attachments), len(testAttachments))
		}
		for i, expected := range testAttachments {
			actual := attachments[i]
			if actual.Info().Size != 13 {
				t.Errorf("got size %d, want %d", actual.Info().Size, 13)
			}
			assertAttachmentEquals(t, actual.Info(), &expected)
		}

		const deletedAt = 100
		util.RequireNil(t,
			repo.Delete(context.Background(), attachments[1].Info().Id, deletedAt))
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
			assertAttachmentEquals(t, actual.Info(), &expected)
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
			assertAttachmentEquals(t, actual.Info(), &expected)

			actual2, err := repo.Find(context.Background(), actual.Info().Id)
			util.RequireNil(t, err)
			assertAttachmentEquals(t, actual2.Info(), &expected)
		}
	}
}

func attachmentsFindAll(repo repo.AttachmentRepo) func(*testing.T) {
	return func(t *testing.T) {
		attachments, err := repo.FindAll(context.Background(), false)
		util.RequireNil(t, err)

		if len(attachments) != 0 {
			t.Errorf("got %d attachments, want %d",
				len(attachments), 0)
		}
	}
}

func assertAttachmentEquals(t *testing.T, actual *repo.AttachmentInfo, expected *repo.UnsavedAttachmentInfo) {
	t.Helper()
	if actual.Filename != expected.Filename {
		t.Errorf("got %q, want %q", actual.Filename, expected.Filename)
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
