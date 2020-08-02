package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
	"log"
	"net"
)

type Service interface {
	Serve(context.Context, net.Listener) error
}

func Serve(ctx context.Context, addr string, grpc, other Service) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("Listening at %s", ln.Addr().String())
	mux := cmux.New(ln)
	lnGrpc := mux.Match(cmux.HTTP2HeaderField("content-type",
		"application/grpc"))
	lnOther := mux.Match(cmux.Any())
	var eg errgroup.Group
	eg.Go(func() error {
		if err := mux.Serve(); err != nil {
			select {
			case <-ctx.Done():
				// don't return error while shutting down
			default:
				return fmt.Errorf("serve mux: %v", err)
			}
		}
		return nil
	})
	eg.Go(func() error {
		eg, ctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			ln := newFakeCloseListener(lnGrpc)
			if err := grpc.Serve(ctx, ln); err != nil {
				return fmt.Errorf("serve grpc: %v", err)
			}
			return nil
		})
		eg.Go(func() error {
			ln := newFakeCloseListener(lnOther)
			if err := other.Serve(ctx, ln); err != nil {
				return fmt.Errorf("serve other: %v", err)
			}
			return nil
		})
		err := eg.Wait()
		if closeErr := ln.Close(); err == nil && closeErr != nil {
			return closeErr
		}
		return err
	})
	return eg.Wait()
}

type fakeCloseListener struct {
	net.Listener
	done chan struct{}
}

func newFakeCloseListener(ln net.Listener) *fakeCloseListener {
	return &fakeCloseListener{
		Listener: ln,
		done:     make(chan struct{}),
	}
}

func (l *fakeCloseListener) Accept() (net.Conn, error) {
	type ret struct {
		conn net.Conn
		err  error
	}
	ch := make(chan ret)
	go func() {
		defer close(ch)
		conn, err := l.Listener.Accept()
		ch <- ret{conn, err}
	}()
	select {
	case ret := <-ch:
		return ret.conn, ret.err
	case <-l.done:
		return nil, &net.OpError{
			Op:     "accept",
			Net:    "tcp",
			Source: nil,
			Addr:   l.Listener.Addr(),
			Err:    errors.New("use of closed network connection"),
		}
	}
}

func (l *fakeCloseListener) Close() error {
	close(l.done)
	return nil
}
