package suite

import (
	"context"
	"log"
	"net"
	"sync"
)

type Server struct {
	wg sync.WaitGroup
	ln net.Listener
}

func Serve(addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := &Server{ln: ln}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.listen(); err != nil {
			log.Println(err)
		}
	}()
	return s, nil
}

func (s *Server) Close() error {
	if err := s.ln.Close(); err != nil {
		return err
	}
	s.wg.Wait()
	return nil
}

func (s *Server) listen() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			if err := handleConn(ctx, conn); err != nil {
				log.Println(err)
			}
		}()
	}
	return err
}
