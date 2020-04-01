package route

import (
	"github.com/gorilla/mux"
	"github.com/tmazeika/testpass/persist"
	"net/http"
)

type res struct {
	router *mux.Router
	db     *persist.DB
}

func Router(db *persist.DB) *mux.Router {
	r := mux.NewRouter()
	res := &res{r, db}

	r.HandleFunc("/attachments/{attachmentId}", res.attachmentHandler).
		Methods(http.MethodGet, http.MethodDelete).
		Name("attachment")
	r.HandleFunc("/attachments", res.attachmentsHandler).
		Methods(http.MethodGet, http.MethodPost)

	return r
}
