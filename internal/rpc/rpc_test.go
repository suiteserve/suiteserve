package rpc

import (
	"context"
	"github.com/stretchr/testify/require"
	pb "github.com/suiteserve/protocol/go/protocol"
	"github.com/suiteserve/suiteserve/internal/repo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
)

const bufSize = 1 << 20
const maxUploadMb = 1

func newClientConn(t *testing.T, r *repo.Repo) grpc.ClientConnInterface {
	ln := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	pb.RegisterQueryServiceServer(srv, &query{
		Repo: r,
	})
	pb.RegisterSuiteServiceServer(srv, &suite{
		Repo:        r,
		maxUploadMb: maxUploadMb,
	})
	go func() {
		require.Nil(t, srv.Serve(ln))
	}()
	t.Cleanup(srv.GracefulStop)

	conn, err := grpc.Dial("bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return ln.Dial()
		}),
		grpc.WithInsecure())
	require.Nil(t, err)
	t.Cleanup(func() {
		require.Nil(t, conn.Close())
	})
	return conn
}
