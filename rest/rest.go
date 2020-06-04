package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/event"
	"github.com/tmazeika/testpass/repo"
	"net/http"
)

type srv struct {
	repos  repo.Repos
	events event.Bus
	router *mux.Router
}

func newSrv(repos repo.Repos) *srv {
	return &srv{
		repos: repos,
		router: mux.NewRouter(),
	}
}

func (s *srv) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func Handler(repos repo.Repos, publicDir string) http.Handler {
	srv := newSrv(repos)

	// middleware
	srv.router.Use(methodOverrideMiddleware)
	srv.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	srv.router.Use(loggingMiddleware)
	srv.router.Use(defaultSecureHeadersMiddleware)

	// API v1
	apiRouter := srv.router.PathPrefix("/v1/").Subrouter()
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
	go srv.consumeRepoChanges()

	// frontend
	publicSrv := http.FileServer(http.Dir(publicDir))
	frontendRouter := srv.router.PathPrefix("/").Subrouter()
	frontendRouter.Use(frontendSecureHeadersMiddleware)
	frontendRouter.PathPrefix("/").Handler(publicSrv)

	return srv
}

func writeJson(w http.ResponseWriter, code int, msg interface{}) error {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	b, err := json.Marshal(&msg)
	if err != nil {
		return fmt.Errorf("marshal json: %v", err)
	}
	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("write json: %v", err)
	}
	return nil
}
