package main

import (
	"flag"
	"github.com/suiteserve/suiteserve/config"
	"github.com/suiteserve/suiteserve/internal/api"
	"github.com/suiteserve/suiteserve/internal/repo"
	"github.com/suiteserve/suiteserve/internal/rpc"
	"log"
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

	if *helpFlag {
		flag.PrintDefaults()
		return
	}
	if *debugFlag {
		log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
		log.Print("Debug mode enabled")
	}

	log.Printf("Using config at %q", *configFileFlag)
	cfg, err := config.Load(*configFileFlag)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	r, err := repo.Open(cfg.Storage.Db)
	if err != nil {
		log.Fatalf("open repo: %v", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("close repo: %v", err)
		}
	}()

	rpcService := rpc.New(cfg.Storage.UserContent.MaxSizeMb, r)
	defer rpcService.Stop()

	srv, err := api.Serve(api.Options{
		Host:                cfg.Http.Host,
		Port:                cfg.Http.Port,
		TlsCertFile:         cfg.Http.TlsCertFile,
		TlsKeyFile:          cfg.Http.TlsKeyFile,
		PublicDir:           cfg.Http.PublicDir,
		UserContentHost:     cfg.Http.UserContentHost,
		UserContentDir:      cfg.Storage.UserContent.Dir,
		UserContentMetaRepo: nil, // TODO
		Rpc:                 rpcService,
	})
	if err != nil {
		log.Fatalf("start http: %v", err)
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	select {
	case <-sigint:
		log.Print("Stopping...")
	case err := <-srv.Err():
		log.Printf("serve http: %v", err)
	}
	srv.Stop()
}
