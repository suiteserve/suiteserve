package handlers

import (
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/database"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"
)

func (s *srv) attachmentHandler(res http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["attachment_id"]
	if !ok {
		panic("request parameter 'attachment_id' not found")
	}

	switch req.Method {
	case http.MethodGet:
		attachment, err := s.db.Attachment(id)
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
		res.Header().Set("Content-Size", strconv.FormatInt(attachment.Size, 10))
		res.Header().Set("Content-Disposition", "inline; filename="+
			strconv.Quote(attachment.Name))
		res.Header().Set("Content-Type", attachment.ContentType)
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
		attachments, err := s.db.AllAttachments()
		if err != nil {
			log.Printf("failed to get attachments: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		httpJson(res, attachments, http.StatusOK)
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

		id, err := s.db.NewAttachment(header.Filename, contentType, src)
		if err != nil {
			log.Printf("failed to save attachment: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		loc, err := s.router.Get("attachment").URL("attachment_id", id)
		if err != nil {
			log.Printf("failed to build URL to attachment: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		res.Header().Set("Location", loc.String())
		httpJson(res, bson.M{"id": id}, http.StatusCreated)
	case http.MethodDelete:
		if err := s.db.DeleteAllAttachments(); err != nil {
			log.Printf("failed to delete attachments: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusNoContent)
	}
}
