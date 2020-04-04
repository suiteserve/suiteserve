package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/database"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"
)

func (s *srv) attachmentHandler(res http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["attachmentId"]

	switch req.Method {
	case http.MethodGet:
		attachment, err := s.db.GetAttachment(id)
		if err == database.ErrNotFound {
			httpError(res, errAttachmentNotFound, http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to get attachment: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		if err := attachment.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			log.Printf("failed to set attachment download deadline: %v\n", err)
			http.Error(res, errUnknown, http.StatusInternalServerError)
			return
		}

		res.Header().Set("Cache-Control", "private, max-age=31536000")
		res.Header().Set("Content-Length", strconv.FormatInt(attachment.Length, 10))
		res.Header().Set("Content-Disposition", "inline; filename="+
			strconv.Quote(attachment.Name))
		res.Header().Set("Content-Type", attachment.Metadata.ContentType)
		res.Header().Set("X-Content-Type-Options", "nosniff")

		if _, err := io.Copy(res, attachment); err != nil {
			log.Printf("failed to copy attachment to response: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		if err := s.db.DeleteAttachment(id); err == database.ErrNotFound {
			httpError(res, errAttachmentNotFound, http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("failed to delete attachment: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusNoContent)
	}
}

func (s *srv) attachmentsHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
	case http.MethodPost:
		src, header, err := req.FormFile("file")
		if err == http.ErrMissingFile {
			httpError(res, errNoAttachmentFile, http.StatusBadRequest)
			return
		} else if err != nil {
			log.Printf("failed to get form file: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := src.Close(); err != nil {
				log.Printf("failed to close form file: %v\n", err)
			}
		}()

		contentType := header.Header.Get("Content-Type")
		if contentType != "" {
			contentType, _, err = mime.ParseMediaType(contentType)
			if err != nil {
				contentType = ""
			}
		}

		id, err := s.db.SaveAttachment(header.Filename, contentType, src)
		if err != nil {
			log.Printf("failed to save attachment: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		loc, err := s.router.Get("attachment").URL("attachmentId", id)
		if err != nil {
			log.Printf("failed to build URL to attachment: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		res.Header().Set("Location", loc.String())
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("X-Content-Type-Options", "nosniff")
		res.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(res).Encode(struct {
			Id string `json:"id"`
		}{id}); err != nil {
			log.Printf("failed to encode JSON: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
	}
}
