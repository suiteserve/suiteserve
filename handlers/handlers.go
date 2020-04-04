package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/database"
	"log"
	"net/http"
	"time"
)

const timeout = 10 * time.Second

const (
	errUnknown = "unknown"

	errNoAttachmentFile   = "no_attachment_file"
	errAttachmentNotFound = "attachment_not_found"
)

type srv struct {
	router *mux.Router
	db     *database.Database
}

func Handler(db *database.Database) http.Handler {
	router := mux.NewRouter()
	srv := &srv{router, db}

	// Serve static files.
	publicSrv := http.FileServer(http.Dir("public/"))
	router.Path("/").Handler(publicSrv)
	router.Path("/favicon.ico").Handler(publicSrv)
	router.PathPrefix("/static/").Handler(publicSrv)

	router.Path("/attachments/{attachmentId}").
		HandlerFunc(srv.attachmentHandler).
		Methods(http.MethodGet, http.MethodDelete).
		Name("attachment")
	router.Path("/attachments").
		HandlerFunc(srv.attachmentsHandler).
		Methods(http.MethodGet, http.MethodPost)

	return methodOverrideHandler(router)
}

func methodOverrideHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			m := req.FormValue("_method")
			if m == http.MethodPut || m == http.MethodPatch || m == http.MethodDelete {
				req.Method = m
			}
		}
		h.ServeHTTP(res, req)
	})
}

func httpError(w http.ResponseWriter, error string, code int) {
	errorJson, err := json.Marshal(struct {
		Error string `json:"error"`
	}{error})
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	_, err = fmt.Fprintln(w, string(errorJson))
	if err != nil {
		log.Printf("failed to write error: %v\n", err)
	}
}
