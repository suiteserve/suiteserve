package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/tmazeika/testpass/database"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	publicDir = "frontend/dist/"
	timeout   = 1 * time.Second
)

type srv struct {
	db         *database.Database
	eventBus   *eventBus
	router     *mux.Router
	wsUpgrader *websocket.Upgrader
}

func Handler(db *database.Database) http.Handler {
	router := mux.NewRouter()
	srv := &srv{
		db,
		&eventBus{
			subscribers: make([]chan event, 0),
		},
		router,
		&websocket.Upgrader{},
	}
	publicSrv := http.FileServer(http.Dir(publicDir))

	router.Use(methodOverrideMiddleware)
	router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	router.Use(loggingMiddleware)
	router.Use(defaultSecureHeadersMiddleware)

	// Static.
	router.PathPrefix("/static/").Handler(publicSrv)
	router.Path("/favicon.ico").Handler(publicSrv)

	// Frontend.
	frontendRouter := router.Path("/").Subrouter()
	frontendRouter.Use(frontendSecureHeadersMiddleware)
	frontendRouter.Path("/").Handler(publicSrv)

	// Attachments.
	router.Path("/attachments/{attachment_id}").
		Handler(errorHandler(srv.attachmentHandler)).
		Methods(http.MethodGet, http.MethodDelete).
		Name("attachment")
	router.Path("/attachments").
		Handler(errorHandler(srv.attachmentCollectionHandler)).
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete)

	// Suites.
	router.Path("/suites/{suite_id}").
		Handler(errorHandler(srv.suiteHandler)).
		Methods(http.MethodGet, http.MethodPatch, http.MethodDelete).
		Name("suite")
	router.Path("/suites").
		Handler(errorHandler(srv.suiteCollectionHandler)).
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete)

	// Cases.
	router.Path("/cases/{case_id}").
		Handler(errorHandler(srv.caseHandler)).
		Methods(http.MethodGet, http.MethodPatch).
		Name("case")
	router.Path("/suites/{suite_id}/cases").
		Handler(errorHandler(srv.caseCollectionHandler)).
		Methods(http.MethodGet, http.MethodPost)

	// Logs.
	router.Path("/logs/{log_id}").
		Handler(errorHandler(srv.logHandler)).
		Methods(http.MethodGet).
		Name("log")
	router.Path("/cases/{case_id}/logs").
		Handler(errorHandler(srv.logCollectionHandler)).
		Methods(http.MethodGet, http.MethodPost)

	// Events.
	router.Path("/events").
		HandlerFunc(srv.eventsHandler).
		Methods(http.MethodGet)

	err := db.WithContext(context.Background()).Watch(func(change database.Change) {
		srv.eventBus.publish(event(change))
	})
	if err != nil {
		log.Fatalln(err)
	}

	return router
}

func methodOverrideMiddleware(h http.Handler) http.Handler {
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

func loggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.URL.String())
		h.ServeHTTP(res, req)
	})
}

func defaultSecureHeadersMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("content-security-policy", "sandbox; "+
			"default-src 'none'; "+
			"base-uri 'none'; "+
			"form-action 'none'; "+
			"frame-ancestors 'none'; "+
			"img-src 'self';")
		res.Header().Set("strict-transport-security", "max-age=31536000")
		res.Header().Set("x-content-type-options", "nosniff")
		res.Header().Set("x-frame-options", "deny")
		h.ServeHTTP(res, req)
	})
}

func frontendSecureHeadersMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("content-security-policy", "block-all-mixed-content; "+
			"default-src 'none'; "+
			"base-uri 'none'; "+
			"connect-src 'self'; "+
			"font-src https://fonts.gstatic.com; "+
			"form-action 'self'; "+
			"frame-ancestors 'none'; "+
			"img-src 'self'; "+
			"script-src 'self' 'unsafe-eval' https://cdn.jsdelivr.net; "+
			"style-src 'self' https://fonts.googleapis.com;")
		res.Header().Set("x-xss-protection", "1; mode=block")
		h.ServeHTTP(res, req)
	})
}

type oneArgHandlerMap map[string]func(http.ResponseWriter, *http.Request, string) error

func (m oneArgHandlerMap) handle(res http.ResponseWriter, req *http.Request, param string) error {
	arg, ok := mux.Vars(req)[param]
	if !ok {
		log.Panicf("req param '%s' not found\n", param)
	}
	fn, ok := m[req.Method]
	if !ok {
		log.Panicf("method %s not implemented\n", req.Method)
	}
	return fn(res, req, arg)
}

type noArgHandlerMap map[string]func(http.ResponseWriter, *http.Request) error

func (m noArgHandlerMap) handle(res http.ResponseWriter, req *http.Request) error {
	fn, ok := m[req.Method]
	if !ok {
		log.Panicf("method %s not implemented\n", req.Method)
	}
	return fn(res, req)
}

func writeJson(res http.ResponseWriter, code int, msg interface{}) error {
	res.Header().Set("content-type", "application/json")
	res.WriteHeader(code)

	if err := json.NewEncoder(res).Encode(msg); err != nil {
		return fmt.Errorf("write json: %v", err)
	}
	return nil
}
