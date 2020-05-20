package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/tmazeika/testpass/repo"
	"net/http"
	"time"
)

const (
	publicDir = "frontend/dist/"
	timeout   = 1 * time.Second
)

type srv struct {
	repos      repo.Repos
	eventBus   *eventBus
	router     *mux.Router
	wsUpgrader *websocket.Upgrader
}

func Handler(repos repo.Repos) http.Handler {
	router := mux.NewRouter()
	srv := &srv{
		repos,
		&eventBus{
			subscribers: make([]chan event, 0),
		},
		router,
		&websocket.Upgrader{},
	}

	router.Use(methodOverrideMiddleware)
	router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	router.Use(loggingMiddleware)
	router.Use(defaultSecureHeadersMiddleware)

	// API v1.
	v1ApiRouter := router.PathPrefix("/v1/").Subrouter()
	// Attachments.
	v1ApiRouter.Path("/attachments/{id}").
		Handler(srv.getAttachmentHandler()).
		Methods(http.MethodGet).
		Name("attachment")
	v1ApiRouter.Path("/attachments/{id}").
		Handler(srv.deleteAttachmentHandler()).
		Methods(http.MethodDelete)
	v1ApiRouter.Path("/attachments").
		Handler(srv.getAttachmentCollectionHandler()).
		Methods(http.MethodGet)
	v1ApiRouter.Path("/attachments").
		Handler(srv.deleteAttachmentCollectionHandler()).
		Methods(http.MethodDelete)
	// Suites.
	v1ApiRouter.Path("/suites/{id}").
		Handler(srv.getSuiteHandler()).
		Methods(http.MethodGet).
		Name("suite")
	v1ApiRouter.Path("/suites/{id}").
		Handler(srv.deleteSuiteHandler()).
		Methods(http.MethodDelete)
	v1ApiRouter.Path("/suites").
		Handler(srv.getSuiteCollectionHandler()).
		Methods(http.MethodGet)
	v1ApiRouter.Path("/suites").
		Handler(srv.deleteSuiteCollectionHandler()).
		Methods(http.MethodDelete)
	// Cases.
	v1ApiRouter.Path("/cases/{id}").
		Handler(srv.getCaseHandler()).
		Methods(http.MethodGet).
		Name("case")
	v1ApiRouter.Path("/suites/{suite_id}/cases").
		Handler(srv.getCaseCollectionHandler()).
		Methods(http.MethodGet)
	// Logs.
	v1ApiRouter.Path("/logs/{id}").
		Handler(srv.getLogHandler()).
		Methods(http.MethodGet).
		Name("log")
	v1ApiRouter.Path("/cases/{case_id}/logs").
		Handler(srv.getLogCollectionHandler()).
		Methods(http.MethodGet)
	// Events.
	v1ApiRouter.Path("/events").
		HandlerFunc(srv.eventsHandler).
		Methods(http.MethodGet)
	go func() {
		for {
			// TODO: handle!
			<-repos.Changes()
		}
	}()

	// Frontend.
	publicSrv := http.FileServer(http.Dir(publicDir))
	frontendRouter := router.PathPrefix("/").Subrouter()
	frontendRouter.Use(frontendSecureHeadersMiddleware)
	frontendRouter.PathPrefix("/").Handler(publicSrv)

	return router
}

func writeJson(w http.ResponseWriter, code int, msg interface{}) error {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(&msg); err != nil {
		return fmt.Errorf("write json: %v", err)
	}
	return nil
}
