package handlers

import (
	"errors"
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
		log.Panicln("req param 'attachment_id' not found")
	}

	switch req.Method {
	case http.MethodGet:
		s.getAttachmentHandler(res, req, id)
	case http.MethodDelete:
		s.deleteAttachmentHandler(res, req, id)
	default:
		log.Panicf("method '%s' not implemented\n", req.Method)
	}
}

func (s *srv) getAttachmentHandler(res http.ResponseWriter, req *http.Request, id string) {
	attachment, err := s.db.WithContext(req.Context()).Attachment(id)
	if errors.Is(err, database.ErrNotFound) {
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

	attachmentFile, err := attachment.OpenFile()
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
}

func (s *srv) deleteAttachmentHandler(res http.ResponseWriter, req *http.Request, id string) {
	if err := s.db.WithContext(req.Context()).DeleteAttachment(id); errors.Is(err, database.ErrNotFound) {
		httpError(res, errNotFound, http.StatusNotFound)
	} else if err != nil {
		log.Printf("delete attachment: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
	} else {
		res.WriteHeader(http.StatusNoContent)
	}
}

func (s *srv) attachmentCollectionHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		s.getAttachmentCollectionHandler(res, req)
	case http.MethodPost:
		s.postAttachmentCollectionHandler(res, req)
	case http.MethodDelete:
		s.deleteAttachmentCollectionHandler(res, req)
	default:
		log.Panicf("method '%s' not implemented\n", req.Method)
	}
}

func (s *srv) getAttachmentCollectionHandler(res http.ResponseWriter, req *http.Request) {
	attachments, err := s.db.WithContext(req.Context()).AllAttachments()
	if err != nil {
		log.Printf("get all attachments: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
		return
	}
	httpJson(res, attachments, http.StatusOK)
}

func (s *srv) postAttachmentCollectionHandler(res http.ResponseWriter, req *http.Request) {
	src, header, err := req.FormFile("file")
	if err != nil {
		httpError(res, errBadFile, http.StatusBadRequest)
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

	id, err := s.db.WithContext(req.Context()).NewAttachment(header.Filename, contentType, src)
	if err != nil {
		log.Printf("new attachment: %v\n", err)
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
}

func (s *srv) deleteAttachmentCollectionHandler(res http.ResponseWriter, req *http.Request) {
	if err := s.db.WithContext(req.Context()).DeleteAllAttachments(); err != nil {
		log.Printf("delete all attachments: %v\n", err)
		httpError(res, errUnknown, http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusNoContent)
}
