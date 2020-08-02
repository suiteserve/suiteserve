package api

import (
	"context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"time"
)

const shutdownTimeout = 3 * time.Second

type HttpOptions struct {
	GrpcServer *grpc.Server

	TlsCertFile string
	TlsKeyFile  string
	PublicDir   string

	UserContentHost     string
	UserContentDir      string
	UserContentMetaRepo UserContentMetaRepo
}

func (o HttpOptions) newHandler() http.Handler {
	var mux http.ServeMux
	mux.Handle("/",
		newGrpcMiddleware(o.GrpcServer,
			newGetMiddleware(
				newSecurityMiddleware(newUiSecurityMiddleware(
					http.FileServer(http.Dir(o.PublicDir)))))))
	mux.Handle(o.UserContentHost+"/",
		newGetMiddleware(
			newSecurityMiddleware(newUserContentSecurityMiddleware(
				newUserContentMiddleware(o.UserContentMetaRepo,
					http.FileServer(http.Dir(o.UserContentDir)))))))
	return newLoggingMiddleware(&mux)
}

type HttpService struct {
	opts HttpOptions
	srv  http.Server
}

func NewHttpService(opts HttpOptions) *HttpService {
	s := HttpService{opts: opts}
	s.srv.Handler = opts.newHandler()
	return &s
}

func (s *HttpService) Serve(ctx context.Context, ln net.Listener) error {
	s.srv.BaseContext = func(net.Listener) context.Context {
		return ctx
	}
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		err := s.srv.ServeTLS(ln, s.opts.TlsCertFile, s.opts.TlsKeyFile)
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	eg.Go(func() error {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(),
			shutdownTimeout)
		defer cancel()
		return s.srv.Shutdown(ctx)
	})
	return eg.Wait()
}
