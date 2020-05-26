package rest

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func (s *srv) getAttachmentHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			panic(ok)
		}
		download, ok := mux.Vars(r)["download"]
		if !ok {
			download = "false"
		}

		attachment, err := s.repos.Attachments().Find(r.Context(), id)
		if err != nil {
			return fmt.Errorf("get attachment: %v", err)
		}

		if download == "false" {
			return writeJson(w, http.StatusOK, attachment)
		} else if attachment.Info().Deleted {
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

func (s *srv) getAttachmentDownloadHandler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			panic(ok)
		}
		download, ok := mux.Vars(r)["download"]
		if !ok {
			download = "false"
		}

		attachment, err := s.repos.Attachments().Find(r.Context(), id)
		if err != nil {
			return fmt.Errorf("get attachment: %v", err)
		}

		if download == "false" {
			return writeJson(w, http.StatusOK, attachment)
		} else if attachment.Info().Deleted {
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

		err := s.repos.Attachments().
			Delete(r.Context(), id, time.Duration(time.Now().UnixNano()).Milliseconds())
		if err != nil {
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
		err := s.repos.Attachments().
			DeleteAll(r.Context(), time.Duration(time.Now().UnixNano()).Milliseconds())
		if err != nil {
			return fmt.Errorf("delete all attachments: %v", err)
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}
