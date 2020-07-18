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

	Rpc Middleware
}

type Server struct {
	addr string
	srv  http.Server
	rpc  Middleware

	err  chan error
	wg   sync.WaitGroup
	once sync.Once
}

func Serve(opts Options) (*Server, error) {
	s := Server{
		rpc: opts.Rpc,
		err: make(chan error),
	}
	s.srv.Addr = net.JoinHostPort(opts.Host, opts.Port)
	s.setSrvHandler(&opts)

	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return nil, err
	}
	s.addr = ln.Addr().String()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		log.Printf("Starting HTTP @ %s", s.addr)
		err := s.srv.ServeTLS(ln, opts.TlsCertFile, opts.TlsKeyFile)
		if err != http.ErrServerClosed {
			log.Printf("serve http: %v", err)
			s.err <- err
		}
		if err := ln.Close(); err != nil {
			log.Printf("close http: %v", err)
		}
		close(s.err)
	}()
	return &s, nil
}

func (s *Server) setSrvHandler(opts *Options) {
	var mux http.ServeMux
	mux.Handle("/",
		s.rpc.NewMiddleware(
			newGetMiddleware(
				newSecurityMiddleware(
					newFrontendSecurityMiddleware(
						http.FileServer(http.Dir(opts.PublicDir)))))))
	mux.Handle(opts.UserContentHost+"/",
		newGetMiddleware(
			newSecurityMiddleware(
				newUserContentSecurityMiddleware(
					newUserContentMiddleware(opts.UserContentMetaRepo,
						http.FileServer(http.Dir(opts.UserContentDir)))))))
	s.srv.Handler = newLoggingMiddleware(&mux)
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
