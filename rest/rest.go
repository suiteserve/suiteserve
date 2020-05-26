package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/tmazeika/testpass/repo"
	"net/http"
)

type srv struct {
	repos      repo.Repos
	eventBus   *eventBus
	router     *mux.Router
	wsUpgrader *websocket.Upgrader
}

func Handler(repos repo.Repos, publicDir string) http.Handler {
	router := mux.NewRouter()
	srv := &srv{
		repos,
		&eventBus{
			subscribers: make([]chan event, 0),
		},
		router,
		&websocket.Upgrader{},
	}

	// middleware
	router.Use(methodOverrideMiddleware)
	router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	router.Use(loggingMiddleware)
	router.Use(defaultSecureHeadersMiddleware)

	// API v1
	v1Router := router.PathPrefix("/v1/").Subrouter()
	// attachments
	v1Router.Path("/attachments/{id}").
		Handler(srv.getAttachmentHandler()).
		Methods(http.MethodGet).
		Queries("download", "{download:true|false}")
	v1Router.Path("/attachments/{id}").
		Handler(srv.getAttachmentHandler()).
		Methods(http.MethodGet)
	v1Router.Path("/attachments/{id}").
		Handler(srv.deleteAttachmentHandler()).
		Methods(http.MethodDelete)
	v1Router.Path("/attachments").
		Handler(srv.getAttachmentCollectionHandler()).
		Methods(http.MethodGet)
	v1Router.Path("/attachments").
		Handler(srv.deleteAttachmentCollectionHandler()).
		Methods(http.MethodDelete)
	// suites
	v1Router.Path("/suites/{id}").
		Handler(srv.getSuiteHandler()).
		Methods(http.MethodGet)
	v1Router.Path("/suites/{id}").
		Handler(srv.deleteSuiteHandler()).
		Methods(http.MethodDelete)
	v1Router.Path("/suites").
		Handler(srv.getSuiteCollectionHandler()).
		Methods(http.MethodGet)
	v1Router.Path("/suites").
		Handler(srv.deleteSuiteCollectionHandler()).
		Methods(http.MethodDelete)
	// cases
	v1Router.Path("/cases/{id}").
		Handler(srv.getCaseHandler()).
		Methods(http.MethodGet)
	v1Router.Path("/suites/{suite_id}/cases").
		Handler(srv.getCaseCollectionHandler()).
		Methods(http.MethodGet)
	// logs
	v1Router.Path("/logs/{id}").
		Handler(srv.getLogHandler()).
		Methods(http.MethodGet)
	v1Router.Path("/cases/{case_id}/logs").
		Handler(srv.getLogCollectionHandler()).
		Methods(http.MethodGet)
	// events
	v1Router.Path("/events").
		HandlerFunc(srv.eventsHandler).
		Methods(http.MethodGet)
	go func() {
		for {
			// TODO: handle!
			<-repos.Changes()
		}
	}()

	// frontend
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
