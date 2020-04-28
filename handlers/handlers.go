package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/database"
	"log"
	"net/http"
	"strings"
)

const (
	publicDir = "public/"
)

var (
	errBadFile  = errors.New("bad_file")
	errBadJson  = errors.New("bad_json")
	errBadQuery = errors.New("bad_query")
	errNotFound = errors.New("not_found")
	errUnknown  = errors.New("unknown")
)

type srv struct {
	db     *database.Database
	router *mux.Router
}

func Handler(db *database.Database) http.Handler {
	router := mux.NewRouter()
	srv := &srv{db, router}

	router.Use(methodOverrideHandler)
	router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.Header().Set("Content-Security-Policy", "sandbox; "+
				"default-src 'none'; "+
				"base-uri 'none'; "+
				"form-action 'none'; "+
				"frame-ancestors 'none'; "+
				"img-src 'self';")
			h.ServeHTTP(res, req)
		})
	})

	// Static files.
	publicSrv := http.FileServer(http.Dir(publicDir))
	router.Path("/").HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Security-Policy", "block-all-mixed-content; "+
			"default-src 'none'; "+
			"base-uri 'none'; "+
			"form-action 'self'; "+
			"frame-ancestors 'none'; "+
			"img-src 'self'; "+
			"script-src 'self' 'unsafe-eval' https://cdn.jsdelivr.net; "+
			"style-src 'self';")
		publicSrv.ServeHTTP(res, req)
	})
	router.Path("/favicon.ico").Handler(publicSrv)
	router.PathPrefix("/static/").Handler(publicSrv)

	// Attachments.
	router.Path("/attachments/{attachment_id}").
		HandlerFunc(srv.attachmentHandler).
		Methods(http.MethodGet, http.MethodDelete).
		Name("attachment")
	router.Path("/attachments").
		HandlerFunc(srv.attachmentCollectionHandler).
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete)

	// Suites.
	router.Path("/suites/{suite_id}").
		HandlerFunc(srv.suiteHandler).
		// TODO: implement Patch
		Methods(http.MethodGet, http.MethodDelete).
		Name("suite")
	router.Path("/suites").
		HandlerFunc(srv.suiteCollectionHandler).
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete)

	// Cases.
	router.Path("/cases/{case_id}").
		HandlerFunc(srv.caseHandler).
		Methods(http.MethodGet, http.MethodPatch).
		Name("case")
	router.Path("/suites/{suite_id}/cases").
		HandlerFunc(srv.caseCollectionHandler).
		Methods(http.MethodGet, http.MethodPost)

	// Logs.
	router.Path("/logs/{log_id}").
		HandlerFunc(srv.logHandler).
		Methods(http.MethodGet).
		Name("log")
	router.Path("/cases/{case_id}/logs").
		HandlerFunc(srv.logCollectionHandler).
		Methods(http.MethodGet, http.MethodPost)

	return router
}

func methodOverrideHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			m := strings.ToUpper(req.FormValue("_method"))
			if m == http.MethodPut || m == http.MethodPatch || m == http.MethodDelete {
				req.Method = m
			}
		}
		h.ServeHTTP(res, req)
	})
}

func httpError(res http.ResponseWriter, error error, code int) {
	httpJson(res, map[string]interface{}{"error": error.Error()}, code)
}

func httpJson(res http.ResponseWriter, v interface{}, code int) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("X-Content-Type-Options", "nosniff")
	res.WriteHeader(code)

	if err := json.NewEncoder(res).Encode(v); err != nil {
		log.Printf("encode json: %v\n", err)
		res.WriteHeader(http.StatusInternalServerError)
		if _, err := fmt.Fprintf(res, `{"error":"`+errUnknown.Error()+`"}"`); err != nil {
			log.Printf("http response: %v\n", err)
		}
	}
}
