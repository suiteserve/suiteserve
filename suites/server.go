package suites

import (
	"context"
	"fmt"
	pb "github.com/suiteserve/protocol/go/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
)

func Serve(ctx context.Context, addr, tlsCert, tlsKey string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen grpc: %v", err)
	}
	creds, err := credentials.NewServerTLSFromFile(tlsCert, tlsKey)
	if err != nil {
		return fmt.Errorf("read grpc tls: %v", err)
	}
	srv := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterSuiteServiceServer(srv, &handlers{})
	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()
	if err := srv.Serve(ln); err != nil {
		return fmt.Errorf("serve grpc: %v", err)
	}
	return nil
}
