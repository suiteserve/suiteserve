package api

import (
	"errors"
	"io"
	"log"
	"net/rpc/jsonrpc"
)

func (s *Server) initRpc() {
	// TODO
	arith := new(Arith)
	if err := s.rpc.Register(arith); err != nil {
		log.Panic("register rpc 'arith'")
	}
}

func (s *Server) serveRpc(conn io.ReadWriteCloser) {
	s.rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
}

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (t *Arith) Divide(args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}
