package handlers

import (
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/database"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (s *srv) attachmentHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		id, ok := mux.Vars(req)["attachmentId"]
		if !ok {
			http.Error(res, "attachment ID is required", http.StatusBadRequest)
			return
		}

		attachment, err := s.db.GetAttachment(id)
		if err == database.ErrBadId || err == database.ErrNotFound {
			http.Error(res, "attachment not found", http.StatusBadRequest)
			return
		} else if err != nil {
			log.Printf("failed to get attachment: %v\n", err)
			http.Error(res, internalErrorTxt, http.StatusInternalServerError)
			return
		}

		if err := attachment.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			log.Printf("failed to set attachment download deadline: %v\n", err)
			http.Error(res, internalErrorTxt, http.StatusInternalServerError)
			return
		}

		res.Header().Set("Cache-Control", "private, max-age=31536000")
		res.Header().Set("Content-Length", strconv.FormatInt(attachment.Length, 10))
		res.Header().Set("Content-Disposition", "inline; filename="+
			strconv.Quote(attachment.Name))
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")

		if _, err := io.Copy(res, attachment); err != nil {
			log.Printf("failed to copy attachment to response: %v\n", err)
			http.Error(res, internalErrorTxt, http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
	}
}

func (s *srv) attachmentsHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
	case http.MethodPost:
		src, header, err := req.FormFile("file")
		if err != nil {
			log.Printf("failed to get form file: %v\n", err)
			http.Error(res, internalErrorTxt, http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := src.Close(); err != nil {
				log.Printf("failed to close form file: %v\n", err)
			}
		}()

		id, err := s.db.SaveAttachment(header.Filename, src)
		if err != nil {
			log.Printf("failed to save attachment: %v\n", err)
			http.Error(res, internalErrorTxt, http.StatusInternalServerError)
			return
		}

		loc, err := s.router.Get("attachment").URL("attachmentId", id)
		if err != nil {
			log.Printf("failed to build URL to attachment: %v\n", err)
			http.Error(res, internalErrorTxt, http.StatusInternalServerError)
			return
		}

		res.Header().Set("Location", loc.String())
		res.WriteHeader(http.StatusCreated)
	}
}
