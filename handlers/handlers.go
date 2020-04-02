package handlers

import (
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/database"
	"net/http"
	"time"
)

const internalErrorTxt = "internal server error"
const timeout = 10 * time.Second

type srv struct {
	router *mux.Router
	db     *database.Database
}

func Router(db *database.Database) *mux.Router {
	router := mux.NewRouter()
	srv := &srv{router, db}

	// Serve static files.
	publicSrv := http.FileServer(http.Dir("public/"))
	router.Handle("/", publicSrv)
	router.PathPrefix("/static/").Handler(publicSrv)

	router.HandleFunc("/attachments/{attachmentId}", srv.attachmentHandler).
		Methods(http.MethodGet, http.MethodDelete).
		Name("attachment")
	router.HandleFunc("/attachments", srv.attachmentsHandler).
		Methods(http.MethodGet, http.MethodPost)

	return router
}
