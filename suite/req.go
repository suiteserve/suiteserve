package suite

import (
	"encoding/json"
	"io"
	"net"
)

type handler func(*msg) (handler, error)

func readRequests(conn net.Conn, handler handler) error {
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)
	dec.UseNumber()
	for {
		var msgJson interface{}
		if err := dec.Decode(&msgJson); err != nil {
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
		m, err := newMsg(msgJson)
		if err != nil {
			return enc.Encode(err)
		}

		next, err := handler(m)
		if err != nil {
			if err, ok := err.(*msg); ok {
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
		handler = next
	}
}
