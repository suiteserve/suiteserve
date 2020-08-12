package api

import (
	"context"
	"golang.org/x/sync/errgroup"
	"log"
	"net"
	"net/http"
	"time"
)

type Options struct {
	TlsCertFile string
	TlsKeyFile  string
	PublicDir   string

	UserContentHost string
	UserContentDir  string
	UserContentRepo FileMetaRepo

	V1 http.Handler
}

func (o Options) newHandler() http.Handler {
	var mux http.ServeMux
	mux.Handle("/",
		newGetHeadMiddleware(
			newSecMiddleware(
				newUiSecMiddleware(
					http.FileServer(http.Dir(o.PublicDir))))))
	mux.Handle(o.UserContentHost+"/",
		newGetHeadMiddleware(
			newSecMiddleware(
				newUserContentMiddleware(o.UserContentRepo,
					http.FileServer(http.Dir(o.UserContentDir))))))
	mux.Handle("/v1/", http.StripPrefix("/v1", o.V1))
	return newLogMiddleware(&mux)
}

func Serve(ctx context.Context, addr string, opts Options) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("Listening at %s", ln.Addr())

	srv := http.Server{
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Handler: opts.newHandler(),
	}
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		err := srv.ServeTLS(ln, opts.TlsCertFile, opts.TlsKeyFile)
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	eg.Go(func() error {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(),
			3*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	})
	return eg.Wait()
}
