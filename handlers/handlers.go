package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/database"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"time"
)

const timeout = 10 * time.Second

const (
	errUnknown  = "unknown"
	errBadJson  = "bad_json"
	errBadQuery = "bad_query"

	errNoAttachmentFile   = "no_attachment_file"
	errAttachmentNotFound = "attachment_not_found"

	errSuiteRunNotFound = "suite_run_not_found"
)

type srv struct {
	router *mux.Router
	db     *database.Database
}

func Handler(db *database.Database) http.Handler {
	router := mux.NewRouter()
	srv := &srv{router, db}

	// Static files.
	publicSrv := http.FileServer(http.Dir("public/"))
	router.Path("/").Handler(publicSrv)
	router.Path("/favicon.ico").Handler(publicSrv)
	router.PathPrefix("/static/").Handler(publicSrv)

	// Attachments.
	router.Path("/attachments/{attachment_id}").
		HandlerFunc(srv.attachmentHandler).
		Methods(http.MethodGet, http.MethodDelete).
		Name("attachment")
	router.Path("/attachments").
		HandlerFunc(srv.attachmentsHandler).
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete)

	// Suites.
	router.Path("/suites/{suite_id}").
		HandlerFunc(srv.suiteHandler).
		Methods(http.MethodGet, http.MethodDelete).
		Name("suite")
	router.Path("/suites").
		HandlerFunc(srv.suitesHandler).
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete)

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
	httpJson(w, bson.M{"error": error}, code)
}

func httpJson(w http.ResponseWriter, v interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode JSON: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := fmt.Fprintf(w, `{"error":"`+errUnknown+`"}"`); err != nil {
			log.Printf("failed to send HTTP response: %v\n", err)
		}
	}
}
