package rest

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	util "github.com/tmazeika/testpass/internal"
	"github.com/tmazeika/testpass/repo"
	"io"
	"log"
	"net/http"
	"strconv"
)

func (s *srv) getAttachmentHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		attachment, err := s.repos.Attachments().Find(r.Context(), id)
		if err == repo.ErrNotFound {
			return errNotFound(err)
		} else if err != nil {
			return fmt.Errorf("get attachment: %v", err)
		}
		if r.FormValue("download") != "true" {
			return writeJson(w, http.StatusOK, attachment)
		}

		info := attachment.Info()
		if info.Deleted {
			return errNotFound(errors.New(id))
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
			if err := file.Close(); err != nil {
				log.Printf("close attachment: %v\n", err)
			}
		}()

		if _, err := io.Copy(w, file); err != nil {
			return fmt.Errorf("write attachment: %v", err)
		}
		return nil
	})
}

func (s *srv) deleteAttachmentHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		err := s.repos.Attachments().Delete(r.Context(), id, util.NowTimeMillis())
		if err == repo.ErrNotFound {
			return errNotFound(err)
		} else if err != nil {
			return fmt.Errorf("delete attachment: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

func (s *srv) getAttachmentCollectionHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		attachments, err := s.repos.Attachments().FindAll(r.Context(), false)
		if err != nil {
			return fmt.Errorf("get all attachments: %v", err)
		}
		return writeJson(w, http.StatusOK, attachments)
	})
}

func (s *srv) deleteAttachmentCollectionHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		err := s.repos.Attachments().DeleteAll(r.Context(), util.NowTimeMillis())
		if err != nil {
			return fmt.Errorf("delete all attachments: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}
