package main

import (
	"context"
	"github.com/suiteserve/suiteserve/internal/api"
	"github.com/suiteserve/suiteserve/internal/rpc"
	"time"
)

func main() {
	srv := api.Serve(api.Options{
		Host:                "localhost",
		Port:                "8080",
		CertFile:            "config/cert.pem",
		KeyFile:             "config/key.pem",
		PublicDir:           "frontend/dist/",
		UserContentHost:     "localhostusercontent",
		UserContentDir:      "data/usercontent/",
		UserContentMetaRepo: repo{},
		Rpc:                 rpc.New(),
	})
	defer srv.Stop()
	for {
		time.Sleep(1 * time.Second)
	}
}

type repo struct{}

func (repo) UserContentMeta(context.Context, string) (*api.FileMeta, error) {
	return &api.FileMeta{
		Filename:    "test.txt",
		ContentType: "text/plain; charset=utf-8",
	}, nil
}
