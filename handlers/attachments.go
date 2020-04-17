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
			httpError(res, errNotFound, http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("get attachment: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		res.Header().Set("Cache-Control", "private, max-age=31536000")
		res.Header().Set("Content-Size", strconv.FormatInt(attachment.Size, 10))
		res.Header().Set("Content-Disposition", "inline; filename="+
			strconv.Quote(attachment.Filename))
		res.Header().Set("Content-Type", attachment.ContentType)
		res.Header().Set("X-Content-Type-Options", "nosniff")

		attachmentFile, err := attachment.Open()
		if err != nil {
			log.Printf("open attachment: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := attachmentFile.Close(); err != nil {
				log.Printf("close attachment: %v\n", err)
			}
		}()

		if _, err := io.Copy(res, attachmentFile); err != nil {
			log.Printf("copy attachment to response: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		if err := s.db.DeleteAttachment(id); err == database.ErrNotFound {
			httpError(res, errNotFound, http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("delete attachment: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusNoContent)
	}
}

func (s *srv) attachmentsHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		attachments, err := s.db.Attachments()
		if err != nil {
			log.Printf("get attachments: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		httpJson(res, attachments, http.StatusOK)
	case http.MethodPost:
		src, header, err := req.FormFile("file")
		if err == http.ErrMissingFile {
			httpError(res, errNoFile, http.StatusBadRequest)
			return
		} else if err != nil {
			log.Printf("get form file: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := src.Close(); err != nil {
				log.Printf("close form file: %v\n", err)
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
			log.Printf("save attachment: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		loc, err := s.router.Get("attachment").URL("attachment_id", id)
		if err != nil {
			log.Printf("build attachment url: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}

		res.Header().Set("Location", loc.String())
		httpJson(res, bson.M{"id": id}, http.StatusCreated)
	case http.MethodDelete:
		if err := s.db.DeleteAttachments(); err != nil {
			log.Printf("delete attachments: %v\n", err)
			httpError(res, errUnknown, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusNoContent)
	}
}
