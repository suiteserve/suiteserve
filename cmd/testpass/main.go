package main

import (
	"context"
	"flag"
	"github.com/suiteserve/suiteserve/config"
	"github.com/suiteserve/suiteserve/repo"
	"github.com/suiteserve/suiteserve/rest"
	"github.com/suiteserve/suiteserve/seed"
	"github.com/suiteserve/suiteserve/suitesrv"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

type Repo interface {
	rest.Repo
	seed.Repo
	suitesrv.Repo

	Seedable() (bool, error)
	Close() error
}

var (
	configFileFlag = flag.String("config", "config/config.json",
		"The path to the JSON configuration file")
	dbFlag = flag.String("db", "buntdb",
		"The database implementation to use: buntdb, mongodb")
	debugFlag = flag.Bool("debug", false,
		"Whether to print extra debug information with log messages")
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

	fileRepo := repo.FileRepo{Pattern: cfg.Storage.Attachments.FilePattern}
	var r Repo
	switch *dbFlag {
	case "buntdb":
		log.Println("Using BuntDB")
		r, err = repo.OpenBuntDb(cfg.Storage.BuntDb.File, &fileRepo)
	case "mongodb":
		// TODO
		log.Fatalln("MongoDB not yet implemented")
	default:
		log.Fatalf("unknown db %q\n", *dbFlag)
	}
	if err != nil {
		log.Fatalf("open db: %v\n", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Fatalf("close db: %v\n", err)
		}
	}()

	if *seedFlag {
		seedable, err := r.Seedable()
		if err != nil {
			log.Fatalf("check db seedability: %v\n", err)
		}
		if seedable {
			log.Println("Seeding database...")
			if err := seed.Seed(r); err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Println("Not seeding non-empty database")
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
	go startHttp(&wg, cfg, r, done)
	go startSuiteSrv(&wg, cfg, r, done)
	wg.Wait()
}

func startHttp(wg *sync.WaitGroup, cfg *config.Config, repo Repo, done <-chan interface{}) {
	defer wg.Done()
	srv := http.Server{
		Addr:    net.JoinHostPort(cfg.Http.Host, strconv.Itoa(int(cfg.Http.Port))),
		Handler: rest.Handler(repo, cfg.Http.PublicDir),
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatalf("listen http: %v\n", err)
	}
	defer ln.Close()
	log.Println("Bound HTTP server to", ln.Addr())

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

func startSuiteSrv(wg *sync.WaitGroup, cfg *config.Config, repo Repo, done <-chan interface{}) {
	defer wg.Done()
	addr := net.JoinHostPort(cfg.SuiteSrv.Host, strconv.Itoa(int(cfg.SuiteSrv.Port)))
	srv, err := suitesrv.Serve(addr, repo, &suitesrv.ServerOptions{
		Timeout:         secondsToDuration(cfg.Storage.Timeout),
		ReconnectPeriod: secondsToDuration(cfg.SuiteSrv.ReconnectPeriod),
		TlsCertFile:     cfg.SuiteSrv.TlsCertFile,
		TlsKeyFile:      cfg.SuiteSrv.TlsKeyFile,
	})
	if err != nil {
		log.Fatalf("start suite srv: %v\n", err)
	}
	<-done
	if err := srv.Close(); err != nil {
		log.Printf("close suite srv: %v\n", err)
	}
}

func secondsToDuration(seconds int) time.Duration {
	return time.Duration(int64(seconds) * time.Second.Nanoseconds())
}
