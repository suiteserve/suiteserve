package main

import (
	"bytes"
	"context"
	"flag"
	"github.com/suiteserve/suiteserve/internal/api"
	"github.com/suiteserve/suiteserve/internal/config"
	"github.com/suiteserve/suiteserve/internal/repo"
	"io/ioutil"
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
	cfg, err := config.Load(*configFlag)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	r := openRepo(cfg)
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

	apiAddr := net.JoinHostPort(cfg.Http.Host,
		strconv.FormatUint(uint64(cfg.Http.Port), 10))
	opts := api.Options{
		Addr: apiAddr,
		TlsCertFile:     cfg.Http.TlsCertFile,
		TlsKeyFile:      cfg.Http.TlsKeyFile,
		PublicDir:       cfg.Http.PublicDir,
		UserContentHost: cfg.Http.UserContentHost,
		UserContentDir:  cfg.Storage.UserContent.Dir,
		UserContentRepo: nil,
		V1:              api.NewV1Handler(r),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		ch := make(chan os.Signal, 1)
		defer close(ch)
		signal.Notify(ch, os.Interrupt)
		defer signal.Stop(ch)
		<-ch
		cancel()
	}()
	if err := api.Serve(ctx, opts); err != nil {
		log.Fatalf("serve api: %v", err)
	}
}

func openRepo(cfg *config.Config) *repo.Repo {
	addr := net.JoinHostPort(cfg.Storage.MongoDb.Host,
		strconv.FormatUint(uint64(cfg.Storage.MongoDb.Port), 10))
	pass, err := ioutil.ReadFile(cfg.Storage.MongoDb.PassFile)
	if err != nil {
		log.Fatalf("read mongodb pass file: %v", err)
	}
	pass = bytes.TrimSuffix(pass, []byte{'\n'})
	log.Printf("Using database at %q", addr)
	r, err := repo.Open(addr, cfg.Storage.MongoDb.ReplSet,
		cfg.Storage.MongoDb.User, string(pass), cfg.Storage.MongoDb.Db)
	if err != nil {
		log.Fatalf("open repo: %v", err)
	}
	return r
}
