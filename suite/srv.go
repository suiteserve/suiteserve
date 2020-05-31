package suite

import (
	"crypto/tls"
	"fmt"
	"github.com/tmazeika/testpass/repo"
	"log"
	"net"
	"sync"
	"time"
)

type Server struct {
	timeout        time.Duration
	wg             sync.WaitGroup
	ln             net.Listener
	suiteRepo      repo.SuiteRepo
	detachedSuites *detachedSuiteStore
	closing        bool
}

type ServerOptions struct {
	Timeout         time.Duration
	ReconnectPeriod time.Duration
	TlsCertFile     string
	TlsKeyFile      string
}

func Serve(addr string, suiteRepo repo.SuiteRepo, opts *ServerOptions) (*Server, error) {
	cert, err := tls.LoadX509KeyPair(opts.TlsCertFile, opts.TlsKeyFile)
	if err != nil {
		return nil, err
	}
	ln, err := tls.Listen("tcp", addr, &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		return nil, err
	}
	log.Println("Bound suite server to", ln.Addr())

	done := make(chan interface{})
	srv := Server{
		timeout:        opts.Timeout,
		ln:             ln,
		suiteRepo:      suiteRepo,
		detachedSuites: newDetachedSuiteStore(opts.ReconnectPeriod, done),
	}
	srv.wg.Add(1)
	go func() {
		defer srv.wg.Done()
		if err := srv.listen(done); err != nil && !srv.closing {
			log.Printf("listen suite srv: %v\n", err)
		}
		close(done)
	}()
	return &srv, nil
}

func (s *Server) Close() error {
	s.closing = true
	if err := s.ln.Close(); err != nil {
		return err
	}
	s.wg.Wait()
	return nil
}

func (s *Server) listen(done <-chan interface{}) error {
	var err error
	for {
		var conn net.Conn
		conn, err = s.ln.Accept()
		if err != nil {
			break
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConn(conn, done)
		}()
	}
	return fmt.Errorf("accept suite srv conn: %v\n", err)
}

func (s *Server) handleConn(conn net.Conn, earlyDone <-chan interface{}) {
	done := make(chan interface{})
	defer close(done)
	go func() {
		select {
		case <-earlyDone:
		case <-done:
		}
		if err := conn.Close(); err != nil {
			log.Printf("close suite srv conn: %v\n", err)
		}
	}()
	handlers := s.newSession(conn)
	if err := readRequests(conn, handlers.hello); err != nil && !s.closing {
		log.Printf("read suite srv conn: %v\n", err)
	}
	if err := handlers.detach(); err != nil {
		log.Printf("detach suite srv conn: %v\n", err)
	}
}
