package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/tmazeika/testpass/config"
	r "gopkg.in/rethinkdb/rethinkdb-go.v6"
	"log"
)

var (
	addrFlag = flag.String("addr", "localhost:28015",
		"The address of the RethinkDB instance to configure")
	adminPassFileFlag = flag.String("admin-pass-file", "config/passfile",
		"The file containing the admin password to use to connect to RethinkDB")
	dbFlag = flag.String("db", "testpass",
		"The name of the database to create for TestPass")
	helpFlag = flag.Bool("help", false,
		"Shows this help")
	tlsCertFileFlag = flag.String("tls-cert-file", "config/cert.pem",
		"The TLS certificate file to use to connect to RethinkDB")
	tlsKeyFileFlag = flag.String("tls-key-file", "config/key.pem",
		"The TLS key file to use to connect to RethinkDB")
	userFlag = flag.String("user", "testpass",
		"The ID of the user to create for TestPass")
	passFileFlag = flag.String("pass-file", "config/passfile",
		"The file containing the password of the user to create for TestPass")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	if *helpFlag {
		flag.PrintDefaults()
		return
	}

	cert, err := tls.LoadX509KeyPair(*tlsCertFileFlag, *tlsKeyFileFlag)
	if err != nil {
		log.Fatalln(err)
	}

	session, err := r.Connect(r.ConnectOpts{
		Address: *addrFlag,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		Username: "admin",
		Password: config.MustReadFile(*adminPassFileFlag),
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = r.DB("rethinkdb").Table("users").Insert(map[string]interface{}{
		"id":       *userFlag,
		"password": config.MustReadFile(*passFileFlag),
	}).Exec(session)
	if err != nil {
		log.Fatalln(err)
	}

	err = r.DBCreate(*dbFlag).Exec(session)
	if err != nil {
		log.Fatalln(err)
	}

	err = r.DB(*dbFlag).Grant(*userFlag, map[string]bool{
		"read":   true,
		"write":  true,
		"config": true,
	}).Exec(session)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Successfully configured RethinkDB")
}
