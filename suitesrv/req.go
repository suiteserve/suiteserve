package suitesrv

import (
	"encoding/json"
	"io"
	"net"
)

type request struct {
	Seq     int64           `json:"seq"`
	Cmd     string          `json:"cmd"`
	Payload json.RawMessage `json:"payload"`
}

type handler func(*request) (handler, error)

func readRequests(conn net.Conn, handler handler) error {
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)
	for {
		var req request
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			if netErr, ok := err.(net.Error); ok {
				if netErr.Temporary() {
					if err := enc.Encode(errTmpIo(netErr.Error())); err != nil {
						return err
					}
					continue
				} else {
					return netErr
				}
			}
			return enc.Encode(errBadJson(err.Error()))
		}

		next, err := handler(&req)
		if err != nil {
			if err, ok := err.(*response); ok {
				if err := enc.Encode(err); err != nil {
					return err
				}
			} else {
				if err := enc.Encode(errOther(0, err)); err != nil {
					return err
				}
			}
			continue
		}
		if next == nil {
			return nil
		}
		handler = next
	}
}
