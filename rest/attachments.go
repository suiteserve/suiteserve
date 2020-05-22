package rest

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *srv) getAttachmentHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]

		download, err := parseBool(r.FormValue("download"))
		if err != nil {
			return errBadQuery(err)
		}
		attachment, err := s.repos.Attachments(r.Context()).Find(id)
		if err != nil {
			return fmt.Errorf("get attachment: %v", err)
		}

		if !*download {
			return writeJson(w, http.StatusOK, attachment)
		} else if attachment.Deleted {
			return errNotFound(errors.New(id))
		}

		panic("nyi")

		//file, err := attachment.OpenFile()
		//if err != nil {
		//	return fmt.Errorf("open attachment: %v", err)
		//}
		//defer func() {
		//	if err := file.Close(); err != nil {
		//		log.Printf("close attachment: %v\n", err)
		//	}
		//}()
		//
		//w.Header().Set("cache-control", "private, max-age=31536000")
		//w.Header().Set("content-size", strconv.FormatInt(attachment.Size, 10))
		//w.Header().Set("content-disposition", "inline; filename="+
		//	strconv.Quote(attachment.Filename))
		//w.Header().Set("content-type", attachment.ContentType)
		//
		//if _, err := io.Copy(w, file); err != nil {
		//	return fmt.Errorf("write attachment: %v", err)
		//}
		//return nil
	})
}

func (s *srv) deleteAttachmentHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]

		err := s.repos.Attachments(r.Context()).Delete(id)
		if err != nil {
			return fmt.Errorf("delete attachment: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

func (s *srv) getAttachmentCollectionHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		attachments, err := s.repos.Attachments(r.Context()).FindAll(false)
		if err != nil {
			return fmt.Errorf("get all attachments: %v", err)
		}
		return writeJson(w, http.StatusOK, attachments)
	})
}

func (s *srv) deleteAttachmentCollectionHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		err := s.repos.Attachments(r.Context()).DeleteAll()
		if err != nil {
			return fmt.Errorf("delete all attachments: %v", err)
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}
