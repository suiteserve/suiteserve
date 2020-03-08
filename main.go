package main

import (
	"flag"
	"git.blazey.dev/tests/auth"
	"git.blazey.dev/tests/handlers"
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
	handlers.Init(router, db)

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
