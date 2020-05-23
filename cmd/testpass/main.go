package main

import (
	"context"
	"flag"
	"github.com/tmazeika/testpass/config"
	"github.com/tmazeika/testpass/repo"
	"github.com/tmazeika/testpass/rest"
	"github.com/tmazeika/testpass/seed"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

var (
	configFileFlag = flag.String("config", "config/config.json",
		"The path to the JSON configuration file")
	dbFlag = flag.String("db", "bunt",
		"The database implementation to use: bunt, mongo")
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

	log.Printf("Using config at %q", *configFileFlag)
	cfg, err := config.New(*configFileFlag)
	if err != nil {
		log.Fatalln(err)
	}

	var repos repo.Repos
	switch *dbFlag {
	case "bunt":
		log.Println("Using BuntDB")
		repos, err = repo.NewBuntRepos(cfg.Storage.Bunt.File, func() string {
			return primitive.NewObjectID().Hex()
		})
		if err != nil {
			log.Fatalf("create BuntDB repos: %v\n", err)
		}
		defer func() {
			if err := repos.Close(); err != nil {
				log.Printf("close BuntDB repos: %v\n", err)
			}
		}()
	case "mongo":
		log.Fatalln("MongoDB not yet implemented")
	default:
		log.Fatalf("unknown db %q\n", *dbFlag)
	}

	if *seedFlag {
		log.Println("Seeding DB...")
		if err := seed.Seed(repos); err != nil {
			log.Fatalln(err)
		}
	}

	listenHttp(cfg, repos)
}

func listenHttp(cfg *config.Config, repos repo.Repos) {
	srv := http.Server{
		Addr:    net.JoinHostPort(cfg.Http.Host, strconv.Itoa(int(cfg.Http.Port))),
		Handler: rest.Handler(repos),
	}
	done := make(chan interface{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		log.Println("Shutting down...")

		ctx, _ := context.WithTimeout(context.Background(), cfg.Http.ShutdownTimeout)
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown http: %v\n", err)
		}
		close(done)
	}()

	log.Println("Binding to", srv.Addr)
	srv.RegisterOnShutdown(func() {
		log.Println("Cleaned up")
	})

	err := srv.ListenAndServeTLS(cfg.Http.TlsCertFile, cfg.Http.TlsKeyFile)
	if err != http.ErrServerClosed {
		log.Fatalf("listen http: %v\n", err)
	}
	<-done
}
