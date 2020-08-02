package main

import (
	"context"
	"flag"
	"github.com/suiteserve/suiteserve/config"
	"github.com/suiteserve/suiteserve/internal/api"
	"github.com/suiteserve/suiteserve/internal/repo"
	"github.com/suiteserve/suiteserve/internal/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	creds, err := credentials.NewServerTLSFromFile(cfg.Http.TlsCertFile,
		cfg.Http.TlsKeyFile)
	if err != nil {
		log.Fatalf("new grpc server: %v", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	grpcWebServer := grpc.NewServer()
	rpc.RegisterServices(grpcServer, cfg.Storage.UserContent.MaxSizeMb, r)
	rpc.RegisterServices(grpcWebServer, cfg.Storage.UserContent.MaxSizeMb, r)
	grpcService := api.NewGrpcService(grpcServer)
	httpService := api.NewHttpService(api.HttpOptions{
		GrpcServer:          grpcWebServer,
		TlsCertFile:         cfg.Http.TlsCertFile,
		TlsKeyFile:          cfg.Http.TlsKeyFile,
		PublicDir:           cfg.Http.PublicDir,
		UserContentHost:     cfg.Http.UserContentHost,
		UserContentDir:      cfg.Storage.UserContent.Dir,
		UserContentMetaRepo: nil, // TODO
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch
		cancel()
	}()
	err = api.Serve(ctx, net.JoinHostPort(cfg.Http.Host, cfg.Http.Port),
		grpcService, httpService)
	if err != nil {
		log.Fatal(err)
	}
}
