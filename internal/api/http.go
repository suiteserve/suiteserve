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

type Service interface {
	http.Handler
	Stop()
}

type Options struct {
	Host      string
	Port      string
	CertFile  string
	KeyFile   string
	PublicDir string

	UserContentHost     string
	UserContentDir      string
	UserContentMetaRepo UserContentMetaRepo

	Rpc Service
}

type Server struct {
	srv http.Server
	rpc Service

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
		defer close(s.err)
		log.Printf("Binding HTTP to %s", net.JoinHostPort(opts.Host, opts.Port))
		err := s.srv.ListenAndServeTLS(opts.CertFile, opts.KeyFile)
		if err != http.ErrServerClosed {
			log.Printf("listen and serve http: %v", err)
			s.err <- err
		}
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
		log.Print("Shutting down HTTP...")
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(),
				shutdownTimeout)
			defer cancel()
			if err := s.srv.Shutdown(ctx); err != nil {
				log.Printf("close http: %v", err)
			}
		}()
		s.rpc.Stop()
		s.wg.Wait()
	})
}
