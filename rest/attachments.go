package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	util "github.com/suiteserve/suiteserve/internal"
	"github.com/suiteserve/suiteserve/repo"
	"io"
	"net/http"
	"strconv"
)

type attachmentFinder interface {
	Attachment(ctx context.Context, id string) (repo.AttachmentFile, error)
	AllAttachments(ctx context.Context) ([]repo.AttachmentFile, error)
}

type attachmentUpdater interface {
	DeleteAttachment(ctx context.Context, id string, at int64) error
	DeleteAllAttachments(ctx context.Context, at int64) error
}

func newGetAttachmentHandler(finder attachmentFinder) http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) (err error) {
		id := mux.Vars(r)["id"]
		attachment, err := finder.Attachment(r.Context(), id)
		if errors.Is(err, repo.ErrNotFound) {
			return errNotFound(err)
		} else if err != nil {
			return fmt.Errorf("get attachment: %v", err)
		}
		if r.FormValue("download") != "true" {
			return writeJson(w, http.StatusOK, attachment)
		}

		info := attachment.Info()
		if info.Deleted {
			return errNotFound(errors.New("deleted"))
		}

		w.Header().Set("cache-control", "private, max-age=31536000")
		w.Header().Set("content-disposition", "inline; filename="+
			strconv.Quote(info.Filename))
		w.Header().Set("content-size",
			strconv.FormatInt(attachment.Info().Size, 10))
		w.Header().Set("content-type", info.ContentType)

		file, err := attachment.Open()
		if err != nil {
			return fmt.Errorf("open attachment: %v", err)
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil && err == nil {
				err = fmt.Errorf("close attachment: %v", err)
			}
		}()

		if _, err := io.Copy(w, file); err != nil {
			return fmt.Errorf("write attachment: %v", err)
		}
		return nil
	})
}

func newDeleteAttachmentHandler(updater attachmentUpdater) http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		err := updater.DeleteAttachment(r.Context(), id, util.NowTimeMillis())
		if errors.Is(err, repo.ErrNotFound) {
			return errNotFound(err)
		} else if err != nil {
			return fmt.Errorf("delete attachment: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

func newGetAttachmentCollectionHandler(finder attachmentFinder) http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		attachments, err := finder.AllAttachments(r.Context())
		if err != nil {
			return fmt.Errorf("get all attachments: %v", err)
		}
		return writeJson(w, http.StatusOK, attachments)
	})
}

func newDeleteAttachmentCollectionHandler(updater attachmentUpdater) http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		err := updater.DeleteAllAttachments(r.Context(), util.NowTimeMillis())
		if err != nil {
			return fmt.Errorf("delete all attachments: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}
