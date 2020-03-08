package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tidwall/buntdb"
	"log"
	"net"
	"net/http"
	"os"
)

type statusError struct {
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

	r := mux.NewRouter()
	srv := srv{db}

	// Logging middleware.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s\n", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	// Content-Type middleware.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			next.ServeHTTP(w, req)
		})
	})

	r.HandleFunc("/users/{name}", func(w http.ResponseWriter, req *http.Request) {
		name := mux.Vars(req)["name"]
		user, err := srv.findUser(name)
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
	}).Methods(http.MethodGet).Name("user")

	r.HandleFunc("/users", func(w http.ResponseWriter, req *http.Request) {
		name := req.FormValue("name")
		pass := req.FormValue("pass")
		role := role(req.FormValue("role"))

		if err := srv.createUser(name, pass, role); err != nil {
			handleErr(w, err)
			return
		}

		loc, err := r.Get("user").URL("name", name)
		if err != nil {
			handleErr(w, err)
			return
		}
		w.Header().Set("Location", loc.String())
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusCreated)
	}).Methods(http.MethodPost)

	r.HandleFunc("/users/{name}", func(w http.ResponseWriter, req *http.Request) {
		name := mux.Vars(req)["name"]
		if err := srv.deleteUser(name); err != nil {
			handleErr(w, err)
			return
		}

		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent)
	}).Methods(http.MethodDelete)

	host := getEnv("HOST", "localhost")
	port := getEnv("PORT", "8080")
	addr := net.JoinHostPort(host, port)
	log.Println("Binding to", addr)
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
	reason := err.Error()
	status := http.StatusInternalServerError
	switch err := err.(type) {
	case statusError:
		status = err.Status
	default:
		log.Println(reason)
	}

	reasonJson, err := json.Marshal(struct {
		Reason string `json:"reason"`
	}{reason})
	if err != nil {
		log.Panicln(err)
	}
	w.WriteHeader(status)
	if _, err := fmt.Fprintln(w, string(reasonJson)); err != nil {
		log.Panicln(err)
	}
}
