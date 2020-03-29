package main

import (
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	r := mux.NewRouter()
	// TODO

	host := config("HOST", "localhost")
	port := config("PORT", "8080")
	addr := net.JoinHostPort(host, port)
	log.Println("Binding to", addr)
	log.Fatalln(http.ListenAndServeTLS(addr, "tls/cert.pem", "tls/key.pem", r))
}

func config(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	} else {
		return def
	}
}
