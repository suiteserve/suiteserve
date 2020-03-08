package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tidwall/buntdb"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	defAdminUser = "admin"
	defAdminPass = "password"
)

type httpError struct {
	error
	Status int
}

type srv struct {
	db *buntdb.DB
}

var (
	createDefAdmin = flag.Bool("defadmin", false,
		"Whether to create the default admin user on startup")
)

func main() {
	flag.Parse()

	dbFile := getEnv("DB_FILE", "data/app.db")
	db, err := buntdb.Open(dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.ReplaceIndex("users.role", "users:*",
		buntdb.IndexJSON("role")); err != nil {
		log.Fatalln(err)
	}

	srv := srv{db}

	var foundAdmin bool
	if err := db.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual("users.role", `{"role":"`+adminRole+`"}`,
			func(key, val string) bool {
				foundAdmin = true
				return false
			})
	}); err != nil {
		log.Fatalln(err)
	}
	if *createDefAdmin || !foundAdmin {
		if *createDefAdmin {
			log.Println("Creating default admin user now")
		} else if !foundAdmin {
			log.Println("No admin user found; creating default now")
		}
		if err := srv.createUser(defAdminUser, defAdminPass, adminRole); err != nil {
			log.Println(err)
		}
	}

	router := mux.NewRouter()

	// Logging middleware.
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s\n", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	// Headers middleware.
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			next.ServeHTTP(w, r)
		})
	})

	router.HandleFunc("/users/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
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

	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		pass := r.FormValue("pass")
		role := role(r.FormValue("role"))

		if err := srv.createUser(name, pass, role); err != nil {
			handleErr(w, err)
			return
		}

		loc, err := router.Get("user").URL("name", name)
		if err != nil {
			handleErr(w, err)
			return
		}
		w.Header().Set("Location", loc.String())
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusCreated)
	}).Methods(http.MethodPost)

	router.HandleFunc("/users/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
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
	log.Fatalln(http.ListenAndServe(addr, router))
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	} else {
		return def
	}
}

func handleErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch err := err.(type) {
	case httpError:
		status = err.Status
	default:
		log.Println(err.Error())
	}

	errorJson, err := json.Marshal(struct {
		Error string `json:"error"`
	}{err.Error()})
	if err != nil {
		log.Panicln(err)
	}
	w.WriteHeader(status)
	if _, err := fmt.Fprintln(w, string(errorJson)); err != nil {
		log.Panicln(err)
	}
}
