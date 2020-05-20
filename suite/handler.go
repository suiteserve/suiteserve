package suite

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

func handleConn(ctx context.Context, conn net.Conn) error {
	dec := json.NewDecoder(conn)
	//enc := json.NewEncoder(conn)

	for {
		var msg map[string]interface{}
		if err := dec.Decode(&msg); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		fmt.Printf("%v\n", msg)
	}
	return nil
}
