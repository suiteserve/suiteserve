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
	apiRouter := router.PathPrefix("/v1/").Subrouter()
	// attachments
	apiRouter.Path("/attachments/{id}").
		Handler(srv.getAttachmentHandler()).
		Methods(http.MethodGet)
	apiRouter.Path("/attachments/{id}").
		Handler(srv.deleteAttachmentHandler()).
		Methods(http.MethodDelete)
	apiRouter.Path("/attachments").
		Handler(srv.getAttachmentCollectionHandler()).
		Methods(http.MethodGet)
	apiRouter.Path("/attachments").
		Handler(srv.deleteAttachmentCollectionHandler()).
		Methods(http.MethodDelete)
	// suites
	apiRouter.Path("/suites/{id}").
		Handler(srv.getSuiteHandler()).
		Methods(http.MethodGet)
	apiRouter.Path("/suites/{id}").
		Handler(srv.deleteSuiteHandler()).
		Methods(http.MethodDelete)
	apiRouter.Path("/suites").
		Handler(srv.getSuiteCollectionHandler()).
		Methods(http.MethodGet)
	apiRouter.Path("/suites").
		Handler(srv.deleteSuiteCollectionHandler()).
		Methods(http.MethodDelete)
	// cases
	apiRouter.Path("/cases/{id}").
		Handler(srv.getCaseHandler()).
		Methods(http.MethodGet)
	apiRouter.Path("/suites/{suite_id}/cases").
		Handler(srv.getCaseCollectionHandler()).
		Methods(http.MethodGet)
	// logs
	apiRouter.Path("/logs/{id}").
		Handler(srv.getLogHandler()).
		Methods(http.MethodGet)
	apiRouter.Path("/cases/{case_id}/logs").
		Handler(srv.getLogCollectionHandler()).
		Methods(http.MethodGet)
	// events
	apiRouter.Path("/events").
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
