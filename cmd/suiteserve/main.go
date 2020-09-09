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
	"strconv"
)

var (
	configFlag = flag.String("config", "config/config.json",
		"The path to the JSON configuration file")
	debugFlag = flag.Bool("debug", false,
		"Whether to print extra debug information with log messages")
	seedFlag = flag.Bool("seed", false,
		"Whether to seed the empty database with sample data")
)

func main() {
	flag.Parse()
	if *debugFlag {
		log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
		log.Print("Debug mode enabled")
	}

	log.Printf("Using config at %q", *configFlag)
	c, err := config.Load(*configFlag)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log.Printf("Using database at %q", "")
	r, err := repo.Open("")
	if err != nil {
		log.Fatalf("open repo: %v", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("close repo: %v", err)
		}
	}()

	if *seedFlag {
		if err := r.Seed(); err != nil {
			log.Fatalf("seed repo: %v", err)
		}
	}

	opts := api.Options{
		Addr: net.JoinHostPort(c.Http.Host,
			strconv.FormatUint(uint64(c.Http.Port), 10)),
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
	if err := api.Serve(ctx, opts); err != nil {
		log.Fatalf("serve api: %v", err)
	}
}
