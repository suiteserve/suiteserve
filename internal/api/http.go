package api

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

const shutdownTimeout = 3 * time.Second

type Options struct {
	Host        string
	Port        string
	TlsCertFile string
	TlsKeyFile  string
	PublicDir   string

	UserContentHost     string
	UserContentDir      string
	UserContentMetaRepo UserContentMetaRepo

	Rpc http.Handler
}

type Server struct {
	srv http.Server
	rpc http.Handler

	err  chan error
	wg   sync.WaitGroup
	once sync.Once
}

func Serve(opts Options) *Server {
	s := Server{
		rpc: opts.Rpc,
		err: make(chan error),
	}
	s.srv.Addr = net.JoinHostPort(opts.Host, opts.Port)
	s.setSrvHandler(&opts)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		log.Printf("Starting HTTP @ %s", net.JoinHostPort(opts.Host, opts.Port))
		err := s.srv.ListenAndServeTLS(opts.TlsCertFile, opts.TlsKeyFile)
		if err != http.ErrServerClosed {
			log.Printf("listen and serve http: %v", err)
			s.err <- err
		}
		close(s.err)
	}()
	return &s
}

func (s *Server) setSrvHandler(opts *Options) {
	var mux http.ServeMux
	mux.Handle("/rpc", s.rpc)
	mux.Handle("/",
		newSecurityMiddleware(
			newFrontendSecurityMiddleware(
				http.FileServer(http.Dir(opts.PublicDir)))))
	mux.Handle(opts.UserContentHost+"/",
		newSecurityMiddleware(
			newUserContentSecurityMiddleware(
				newUserContentMiddleware(opts.UserContentMetaRepo,
					http.FileServer(http.Dir(opts.UserContentDir))))))
	s.srv.Handler = newLoggingMiddleware(newGetMiddleware(&mux))
}

func (s *Server) Err() <-chan error {
	return s.err
}

func (s *Server) Stop() {
	s.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(),
			shutdownTimeout)
		defer cancel()
		if err := s.srv.Shutdown(ctx); err != nil {
			log.Printf("stop http: %v", err)
		}
		s.wg.Wait()
	})
}
