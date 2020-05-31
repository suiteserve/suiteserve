package suite

import (
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
}

type ServerOptions struct {
	Timeout         time.Duration
	ReconnectPeriod time.Duration
}

func Serve(addr string, suiteRepo repo.SuiteRepo, opts *ServerOptions) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	done := make(chan interface{})
	s := Server{
		timeout:        opts.Timeout,
		ln:             ln,
		suiteRepo:      suiteRepo,
		detachedSuites: newDetachedSuiteStore(opts.ReconnectPeriod, done),
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.listen(); err != nil {
			log.Printf("listen suite srv: %v\n", err)
		}
		close(done)
	}()
	return &s, nil
}

func (s *Server) Close() error {
	if err := s.ln.Close(); err != nil {
		return fmt.Errorf("close suite srv: %v\n", err)
	}
	s.wg.Wait()
	return nil
}

func (s *Server) listen() error {
	log.Println("Bound suite srv to", s.ln.Addr())
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
			defer func() {
				if err := conn.Close(); err != nil {
					log.Printf("close suite srv conn: %v\n", err)
				}
			}()
			s.handleConn(conn)
		}()
	}
	return fmt.Errorf("accept suite srv conn: %v\n", err)
}

func (s *Server) handleConn(conn net.Conn) {
	handlers := s.newSession(conn)
	if err := readRequests(conn, handlers.helloHandler); err != nil {
		log.Printf("read suite srv conn: %v\n", err)
	}
	if err := handlers.detach(); err != nil {
		log.Printf("detach suite srv conn: %v\n", err)
	}
}
