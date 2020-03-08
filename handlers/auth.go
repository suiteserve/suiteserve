package handlers

import (
	"encoding/json"
	"fmt"
	"git.blazey.dev/tests/auth"
	"github.com/gorilla/mux"
	"net/http"
)

func (e env) user(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		e.getUser(w, r)
	case http.MethodDelete:
		e.deleteUser(w, r)
	}
}

func (e env) users(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		e.getAllUsers(w, r)
	case http.MethodPost:
		e.createUser(w, r)
	}
}

func (e env) createUser(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	pass := r.FormValue("pass")
	role := auth.Role(r.FormValue("role"))

	if _, err := auth.CreateUser(e.db, name, pass, role); err != nil {
		handleErr(w, err)
		return
	}

	loc, err := e.router.Get("user").URL("name", name)
	if err != nil {
		handleErr(w, err)
		return
	}

	w.Header().Set("Location", loc.String())
	w.Header().Del("Content-Type")
	w.WriteHeader(http.StatusCreated)
}

func (e env) getUser(w http.ResponseWriter, r *http.Request) {
	user, err := auth.FindUserByName(e.db, mux.Vars(r)["name"])
	if err != nil {
		handleErr(w, err)
		return
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		handleErr(w, err)
		return
	}
	if _, err := fmt.Fprintln(w, string(userJson)); err != nil {
		handleErr(w, err)
		return
	}
}

func (e env) getAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := auth.FindAllUsers(e.db)
	if err != nil {
		handleErr(w, err)
		return
	}

	usersJson, err := json.Marshal(users)
	if err != nil {
		handleErr(w, err)
		return
	}
	if _, err := fmt.Fprintln(w, string(usersJson)); err != nil {
		handleErr(w, err)
		return
	}
}

func (e env) deleteUser(w http.ResponseWriter, r *http.Request) {
	if err := auth.DeleteUser(e.db, mux.Vars(r)["name"]); err != nil {
		handleErr(w, err)
		return
	}

	w.Header().Del("Content-Type")
	w.WriteHeader(http.StatusNoContent)
}