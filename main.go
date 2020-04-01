package main

import (
	"github.com/tmazeika/testpass/persist"
	"github.com/tmazeika/testpass/route"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	db, err := persist.New(config("MONGO_HOST", ""), config("MONGO_USER", ""),
		config("MONGO_PASS", ""))
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Println(err)
		}
	}()

	r := route.Router(db)
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
