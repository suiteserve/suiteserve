package main

import (
	"context"
	"flag"
	"github.com/suiteserve/suiteserve/internal/api"
	"github.com/suiteserve/suiteserve/internal/config"
	"github.com/suiteserve/suiteserve/internal/repo"
	"log"
	"net"
	"os"
	"os/signal"
)

var (
	configFileFlag = flag.String("config", "config/config.json",
		"The path to the JSON configuration file")
	debugFlag = flag.Bool("debug", false,
		"Whether to print extra debug information with log messages")
	helpFlag = flag.Bool("help", false,
		"Shows this help")
	seedFlag = flag.Bool("seed", false,
		"Whether to first seed the database with test data")
)

func main() {
	flag.Parse()
	// TODO: not necessary in Go 1.15
	if *helpFlag {
		flag.PrintDefaults()
		return
	}
	if *debugFlag {
		log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
		log.Print("Debug mode enabled")
	}

	log.Printf("Using config at %q", *configFileFlag)
	c, err := config.Load(*configFileFlag)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	r, err := repo.Open(c.Storage.Db)
	if err != nil {
		log.Fatalf("open repo: %v", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("close repo: %v", err)
		}
	}()

	opts := api.Options{
		TlsCertFile:     c.Http.TlsCertFile,
		TlsKeyFile:      c.Http.TlsKeyFile,
		PublicDir:       c.Http.PublicDir,
		UserContentHost: c.Http.UserContentHost,
		UserContentDir:  c.Storage.UserContent.Dir,
		UserContentRepo: nil,
		V1:              api.NewV1Handler(r),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch
		cancel()
	}()
	err = api.Serve(ctx, net.JoinHostPort(c.Http.Host, c.Http.Port), opts)
	if err != nil {
		log.Fatal(err)
	}
}
