package handlers

import (
	"errors"
	"fmt"
	"github.com/tmazeika/testpass/database"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
)

func (s *srv) attachmentHandler(res http.ResponseWriter, req *http.Request) error {
	return oneArgHandlerMap{
		http.MethodGet:    s.getAttachmentHandler,
		http.MethodDelete: s.deleteAttachmentHandler,
	}.handle(res, req, "attachment_id")
}

func (s *srv) getAttachmentHandler(res http.ResponseWriter, req *http.Request, id string) error {
	attachment, err := s.db.WithContext(req.Context()).Attachment(id)
	if errors.Is(err, database.ErrNotFound) {
		return errNotFound
	} else if err != nil {
		return fmt.Errorf("get attachment: %v", err)
	}

	res.Header().Set("cache-control", "private, max-age=31536000")
	res.Header().Set("content-size", strconv.FormatInt(attachment.Size, 10))
	res.Header().Set("content-disposition", "inline; filename="+
		strconv.Quote(attachment.Filename))
	res.Header().Set("content-type", attachment.ContentType)

	file, err := attachment.OpenFile()
	if err != nil {
		return fmt.Errorf("open attachment: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("close attachment: %v\n", err)
		}
	}()

	if _, err := io.Copy(res, file); err != nil {
		return fmt.Errorf("write attachment: %v", err)
	}
	return nil
}

func (s *srv) deleteAttachmentHandler(res http.ResponseWriter, req *http.Request, id string) error {
	err := s.db.WithContext(req.Context()).DeleteAttachment(id)
	if errors.Is(err, database.ErrNotFound) {
		return errNotFound
	} else if err != nil {
		return fmt.Errorf("delete attachment: %v", err)
	}

	res.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *srv) attachmentCollectionHandler(res http.ResponseWriter, req *http.Request) error {
	return noArgHandlerMap{
		http.MethodGet:    s.getAttachmentCollectionHandler,
		http.MethodPost:   s.postAttachmentCollectionHandler,
		http.MethodDelete: s.deleteAttachmentCollectionHandler,
	}.handle(res, req)
}

func (s *srv) getAttachmentCollectionHandler(res http.ResponseWriter, req *http.Request) error {
	attachments, err := s.db.WithContext(req.Context()).AllAttachments()
	if err != nil {
		return fmt.Errorf("get all attachments: %v", err)
	}

	writeJson(res, attachments, http.StatusOK)
	return nil
}

func (s *srv) postAttachmentCollectionHandler(res http.ResponseWriter, req *http.Request) error {
	file, header, err := req.FormFile("file")
	if err == http.ErrMissingFile {
		return errBadFile
	} else if err != nil {
		return fmt.Errorf("get form file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
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

	id, err := s.db.WithContext(req.Context()).NewAttachment(header.Filename, contentType, file)
	if err != nil {
		return fmt.Errorf("new attachment: %v", err)
	}

	loc, err := s.router.Get("attachment").URL("attachment_id", id)
	if err != nil {
		return fmt.Errorf("build attachment url: %v", err)
	}

	res.Header().Set("Location", loc.String())
	writeJson(res, bson.M{"id": id}, http.StatusCreated)
	return nil
}

func (s *srv) deleteAttachmentCollectionHandler(res http.ResponseWriter, req *http.Request) error {
	err := s.db.WithContext(req.Context()).DeleteAllAttachments()
	if err != nil {
		return fmt.Errorf("delete all attachments: %v", err)
	}

	res.WriteHeader(http.StatusNoContent)
	return nil
}
