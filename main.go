package main

import (
	"github.com/tmazeika/testpass/config"
	"github.com/tmazeika/testpass/database"
	"github.com/tmazeika/testpass/handlers"
	"log"
	"net"
	"net/http"
)

func main() {
	db, err := database.Open()
	if err != nil {
		log.Fatalf("failed to open DB: %v\n", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close DB: %v\n", err)
		}
	}()

	handler := handlers.Handler(db)
	host := config.Get(config.Host, "localhost")
	port := config.Get(config.Port, "8080")
	addr := net.JoinHostPort(host, port)

	log.Println("Binding to", addr)
	// TODO: implement proper error handling for ListenAndServeTLS
	log.Fatalln(http.ListenAndServeTLS(addr, "tls/cert.pem", "tls/key.pem", handler))
}
