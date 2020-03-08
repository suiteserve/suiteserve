package handlers

import (
	"encoding/json"
	"fmt"
	"git.blazey.dev/tests/auth"
	"github.com/gorilla/mux"
	"github.com/tidwall/buntdb"
	"log"
	"net/http"
)

type env struct {
	router *mux.Router
	db *buntdb.DB
}

func Init(router *mux.Router, db *buntdb.DB) {
	env := env{router, db}
	usersRouter := router.PathPrefix("/users").Subrouter()

	usersRouter.HandleFunc("/{name}", env.user).
		Methods(http.MethodGet, http.MethodDelete).Name("user")

	usersRouter.HandleFunc("", env.users).
		Methods(http.MethodGet, http.MethodPost)
}

func handleErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch err {
	case auth.ErrUserExists:
		status = http.StatusConflict
	case auth.ErrUserNotFound:
		status = http.StatusNotFound
	default:
		log.Println(err.Error())
	}

	errJson, err := json.Marshal(map[string]string{
		"error": err.Error(),
	})
	if err != nil {
		log.Println(err)
		return
	}
	w.WriteHeader(status)
	if _, err := fmt.Fprintln(w, string(errJson)); err != nil {
		log.Println(err)
	}
}
