package suite

import (
	"encoding/json"
	"net"
)

type request struct {
	seq int64
	obj map[string]interface{}
}

func newRequest(msg map[string]interface{}) (*request, error) {
	seqJson, ok := msg["seq"].(json.Number)
	if !ok {
		return nil, errBadSeq(msg["seq"], "not an int")
	}
	seq, err := seqJson.Int64()
	if err != nil {
		return nil, errBadSeq(seqJson, "not an int")
	}
	if seq < 1 {
		return nil, errBadSeq(seq, "nonpositive")
	}
	delete(msg, "seq")
	return &request{
		seq: seq,
		obj: msg,
	}, nil
}

type handler func(*request) (handler, error)

func readRequests(conn net.Conn, handler handler) error {
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)
	dec.UseNumber()
	for {
		var msg map[string]interface{}
		if err := dec.Decode(&msg); err != nil {
			if err, ok := err.(net.Error); !ok || !err.Temporary() {
				return err
			}
			if err := enc.Encode(errOther(0, err)); err != nil {
				return err
			}
			continue
		}

		next, err := handleMessage(msg, handler)
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
		handler = next
	}
}

func handleMessage(msg map[string]interface{}, handler handler) (handler, error) {
	r, err := newRequest(msg)
	if err != nil {
		return nil, err
	}
	next, err := handler(r)
	if err != nil {
		return nil, err
	}
	return next, nil
}
