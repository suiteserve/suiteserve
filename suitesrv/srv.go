package suitesrv

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type Repo interface {
	suiteInserter
	suiteUpdater
	caseInserter
	caseFinder
	caseUpdater
	logInserter
}

type Server struct {
	timeout         time.Duration
	reconnectPeriod time.Duration
	wg              sync.WaitGroup
	ln              net.Listener
	repo            Repo
	cancel          func()
	closing         bool
}

type ServerOptions struct {
	Timeout         time.Duration
	ReconnectPeriod time.Duration
	TlsCertFile     string
	TlsKeyFile      string
}

func Serve(addr string, repo Repo, opts *ServerOptions) (*Server, error) {
	cert, err := tls.LoadX509KeyPair(opts.TlsCertFile, opts.TlsKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load tls cert and key: %v", err)
	}
	ln, err := tls.Listen("tcp", addr, &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		return nil, fmt.Errorf("listen suite srv: %v\n", err)
	}
	log.Println("Bound suite server to", ln.Addr())

	ctx, cancel := context.WithCancel(context.Background())
	srv := Server{
		timeout:         opts.Timeout,
		reconnectPeriod: opts.ReconnectPeriod,
		ln:              ln,
		repo:            repo,
		cancel:          cancel,
	}
	srv.wg.Add(1)
	go func() {
		defer srv.wg.Done()
		if err := srv.handleConns(ctx); err != nil && !srv.closing {
			log.Printf("handle suite srv conns: %v\n", err)
		}
	}()
	return &srv, nil
}

func (s *Server) Close() error {
	s.closing = true
	s.cancel()
	if err := s.ln.Close(); err != nil {
		return err
	}
	s.wg.Wait()
	return nil
}

func (s *Server) handleConns(ctx context.Context) error {
	var err error
	for {
		var conn net.Conn
		if conn, err = s.ln.Accept(); err != nil {
			break
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleNextConn(ctx, conn)
		}()
	}
	return err
}

func (s *Server) handleNextConn(ctx context.Context, conn net.Conn) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil && !s.closing {
			log.Printf("close suite srv conn: %v\n", err)
		}
	}()
	handlers := s.newSession(ctx, conn)
	err := readRequests(conn, handlers.hello)
	if err != nil && !s.closing {
		log.Printf("read suite srv conn: %v\n", err)
	}
	if err := handlers.disconnect(); err != nil {
		log.Printf("disconnect suite srv conn: %v\n", err)
	}
}
