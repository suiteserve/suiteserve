package main

import (
	"context"
	"flag"
	"github.com/tmazeika/testpass/config"
	"github.com/tmazeika/testpass/repo"
	"github.com/tmazeika/testpass/rest"
	"github.com/tmazeika/testpass/seed"
	"github.com/tmazeika/testpass/suite"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

var (
	configFileFlag = flag.String("config", "config/config.json",
		"The path to the JSON configuration file")
	debugFlag = flag.Bool("debug", false,
		"Whether to print extra debug information with log messages")
	dbFlag = flag.String("db", "bunt",
		"The database implementation to use: bunt, mongo")
	helpFlag = flag.Bool("help", false,
		"Shows this help")
	seedFlag = flag.Bool("seed", false,
		"Whether to first seed the database with test data")
)

func main() {
	flag.Parse()

	if *debugFlag {
		log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)
		log.Println("Debug mode enabled")
	}

	if *helpFlag {
		flag.PrintDefaults()
		return
	}

	log.Printf("Using config at %q", *configFileFlag)
	cfg, err := config.New(*configFileFlag)
	if err != nil {
		log.Fatalln(err)
	}

	var repos repo.Repos
	switch *dbFlag {
	case "bunt":
		log.Println("Using BuntDB")
		repos, err = repo.OpenBuntRepos(cfg.Storage.BuntDb.File,
			cfg.Storage.Attachments.FilePattern, nil)
	case "mongo":
		// TODO
		log.Fatalln("MongoDB not yet implemented")
	default:
		log.Fatalf("unknown db %q\n", *dbFlag)
	}
	if err != nil {
		log.Fatalf("open repos: %v\n", err)
	}
	defer func() {
		if err := repos.Close(); err != nil {
			log.Printf("close repos: %v\n", err)
		}
	}()

	if *seedFlag {
		if repos.StartedEmpty() {
			log.Println("Seeding DB...")
			if err := seed.Seed(repos); err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Println("Not seeding non-empty DB")
		}
	}

	done := make(chan interface{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		log.Println("Shutting down...")
		close(done)
	}()

	var wg sync.WaitGroup
	wg.Add(2)
	go listenHttp(&wg, cfg, repos, done)
	go listenSuiteSrv(&wg, cfg, repos.Suites(), done)
	wg.Wait()
}

func listenHttp(wg *sync.WaitGroup, cfg *config.Config, repos repo.Repos, done <-chan interface{}) {
	defer wg.Done()
	srv := http.Server{
		Addr:    net.JoinHostPort(cfg.Http.Host, strconv.Itoa(int(cfg.Http.Port))),
		Handler: rest.Handler(repos, cfg.Http.PublicDir),
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatalf("listen http: %v\n", err)
	}
	defer func() {
		if err := ln.Close(); err != nil {
			log.Printf("close http: %v\n", err)
		}
	}()
	log.Println("Bound HTTP to", ln.Addr())

	go func() {
		<-done
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Http.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown http: %v\n", err)
		}
	}()

	err = srv.ServeTLS(ln, cfg.Http.TlsCertFile, cfg.Http.TlsKeyFile)
	if err != http.ErrServerClosed {
		log.Fatalf("serve http: %v\n", err)
	}
}

func listenSuiteSrv(wg *sync.WaitGroup, cfg *config.Config, suiteRepo repo.SuiteRepo, done <-chan interface{}) {
	defer wg.Done()
	srv, err := suite.Serve(net.JoinHostPort(cfg.Suite.Host, strconv.Itoa(int(cfg.Suite.Port))), suiteRepo, &suite.ServerOptions{
		Timeout:         secondsToDuration(cfg.Storage.Timeout),
		ReconnectPeriod: secondsToDuration(cfg.Suite.ReconnectPeriod),
	})
	if err != nil {
		log.Fatalln(err)
	}
	<-done
	if err := srv.Close(); err != nil {
		log.Println(err)
	}
}

func secondsToDuration(seconds int) time.Duration {
	return time.Duration(int64(seconds) * time.Second.Nanoseconds())
}
