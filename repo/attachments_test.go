package repo

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
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

func TestBuntRepo(t *testing.T) {
	dir, err := ioutil.TempDir("", "data")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	i := 0
	repos, err := NewBuntRepos(filepath.Join(dir, "bunt.db"), func() string {
		i++
		return strconv.Itoa(i)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer repos.Close()

	go func() {
		for {
			// ignore changes
			<-repos.Changes()
		}
	}()

	t.Run("Attachments", func(t *testing.T) {
		attachments := repos.Attachments(context.Background())

		t.Run("Save_Find", attachmentsSaveFind(attachments))
		t.Run("Find*_Delete*", attachmentsFindDelete(attachments))
	})
}

func attachmentsSaveFind(repo AttachmentRepo) func(*testing.T) {
	return func(t *testing.T) {
		for i, expected := range testAttachments {
			id, err := repo.Save(expected)
			if err != nil {
				t.Fatal(err)
			}
			if id != strconv.Itoa(i+1) {
				t.Errorf("got %q, want %d", id, i+1)
			}

			actual, err := repo.Find(id)
			if err != nil {
				t.Fatal(err)
			}
			if actual.Id != id {
				t.Errorf("got %q, want %q", actual.Id, id)
			}
			assertAttachmentEquals(t, actual, &expected)
		}
	}
}

func attachmentsFindDelete(repo AttachmentRepo) func(*testing.T) {
	return func(t *testing.T) {
		attachments, err := repo.FindAll(true)
		if err != nil {
			t.Fatal(err)
		}

		if len(attachments) != len(testAttachments) {
			t.Errorf("got %d attachments, want %d attachments",
				len(attachments), len(testAttachments))
		}
		for i, expected := range testAttachments {
			actual := attachments[i]
			assertAttachmentEquals(t, &actual, &expected)
		}

		deletedAt := int64(1590186181968)
		if err := repo.Delete(attachments[1].Id, deletedAt); err != nil {
			t.Fatal(err)
		}
		attachments, err = repo.FindAll(true)
		if err != nil {
			t.Fatal(err)
		}
		if len(attachments) != len(testAttachments) {
			t.Errorf("got %d attachments, want %d attachments",
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

		if err := repo.DeleteAll(deletedAt); err != nil {
			t.Fatal(err)
		}
		attachments, err = repo.FindAll(true)
		if err != nil {
			t.Fatal(err)
		}
		if len(attachments) != len(testAttachments) {
			t.Errorf("got %d attachments, want %d attachments",
				len(attachments), len(testAttachments))
		}
		for i, expected := range testAttachments {
			if !expected.Deleted {
				expected.Deleted = true
				expected.DeletedAt = deletedAt
			}
			actual := attachments[i]
			assertAttachmentEquals(t, &actual, &expected)

			actual2, err := repo.Find(actual.Id)
			if err != nil {
				t.Fatal(err)
			}
			assertAttachmentEquals(t, actual2, &expected)
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
