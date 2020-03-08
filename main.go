package main

import (
	"github.com/gorilla/mux"
	"github.com/tidwall/buntdb"
	"log"
	"net"
	"net/http"
	"os"
)

type StatusError struct {
	error
	Status int
}

type srv struct {
	db *buntdb.DB
}

func main() {
	dbFile := getEnv("DB_FILE", "data/app.db")
	db, err := buntdb.Open(dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	srv := srv{db}

	r := mux.NewRouter()
	r.HandleFunc("/users/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		userJson, err := srv.getUserJson(name)
		if err != nil {
			handleErr(w, err)
			return
		}
		if _, err := w.Write([]byte(userJson)); err != nil {
			handleErr(w, err)
			return
		}
	}).Methods(http.MethodGet)

	r.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		pass := r.FormValue("pass")
		role := role(r.FormValue("role"))

		if err := srv.createUser(name, pass, role); err != nil {
			handleErr(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}).Methods(http.MethodPost)

	host := getEnv("HOST", "localhost")
	port := getEnv("PORT", "8080")
	addr := net.JoinHostPort(host, port)
	log.Fatalln(http.ListenAndServe(addr, r))
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	} else {
		return def
	}
}

func handleErr(w http.ResponseWriter, err error) {
	log.Println(err.Error())
	switch e := err.(type) {
	case StatusError:
		http.Error(w, e.Error(), e.Status)
	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
