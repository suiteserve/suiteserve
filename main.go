package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"git.blazey.dev/tests/auth"
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

	// Create default admin user when necessary.
	admins, err := auth.FindUsersByRole(db, auth.AdminRole)
	if err != nil {
		log.Fatalln(err)
	}
	if *createDefAdmin || len(admins) == 0 {
		if *createDefAdmin {
			log.Println("Creating default admin user now")
		} else if len(admins) == 0 {
			log.Println("No admin user found; creating default now")
		}
		if _, err := auth.CreateUser(db, defAdminUser, defAdminPass, auth.AdminRole); err != nil {
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
		user, err := auth.FindUserByName(db, mux.Vars(r)["name"])
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
		role := auth.Role(r.FormValue("role"))

		if _, err := auth.CreateUser(db, name, pass, role); err != nil {
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

	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		users, err := auth.FindAllUsers(db)
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
	}).Methods(http.MethodGet)

	router.HandleFunc("/users/{name}", func(w http.ResponseWriter, r *http.Request) {
		if err := auth.DeleteUser(db, mux.Vars(r)["name"]); err != nil {
			handleErr(w, err)
			return
		}

		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent)
	}).Methods(http.MethodDelete)

	host := getEnv("HOST", "localhost")
	port := getEnv("PORT", "8080")
	addr := net.JoinHostPort(host, port)
	tlsCert := getEnv("TLS_CERT", "tls/cert.pem")
	tlsKey := getEnv("TLS_KEY", "tls/key.pem")

	log.Println("Binding to", addr)
	log.Fatalln(http.ListenAndServeTLS(addr, tlsCert, tlsKey, router))
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
