package main

import (
	"context"
	"github.com/tmazeika/testpass/config"
	"github.com/tmazeika/testpass/database"
	"github.com/tmazeika/testpass/handlers"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// The graceful shutdown timeout.
const timeout = 10 * time.Second

func main() {
	log.Println("Starting up...")
	db, err := database.Open()
	if err != nil {
		log.Fatalf("open db: %v\n", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close db: %v\n", err)
		}
	}()
	listenHttp(db)
}

func listenHttp(db *database.Database) {
	host := config.Get(config.Host, "localhost")
	port := config.Get(config.Port, "8080")
	srv := http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: handlers.Handler(db),
	}
	srvDone := make(chan interface{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		log.Println("Shutting down...")

		ctx, _ := context.WithTimeout(context.Background(), timeout)
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown http: %v\n", err)
		}
		close(srvDone)
	}()

	log.Println("Binding to", srv.Addr)
	if err := srv.ListenAndServeTLS("tls/cert.pem", "tls/key.pem"); err != http.ErrServerClosed {
		log.Fatalf("listen http: %v\n", err)
	}
	<-srvDone
}
