package jsonrpc2

import (
	"encoding/json"
	"errors"
	"io"
	"net/rpc"
	"sync"
)

type serverCodec struct {
	closer io.Closer
	enc    *json.Encoder
	dec    *json.Decoder

	req serverReq

	mu      sync.Mutex
	seq     uint64
	pending map[uint64]json.RawMessage
}

// NewServerCodec returns a new rpc.ServerCodec using JSON-RPC 2.0 on conn.
func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{
		closer:  conn,
		enc:     json.NewEncoder(conn),
		dec:     json.NewDecoder(conn),
		pending: make(map[uint64]json.RawMessage),
	}
}

type serverReq struct {
	JsonRpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Id      json.RawMessage `json:"id"`
}

func (r *serverReq) reset() {
	r.JsonRpc = ""
	r.Method = ""
	r.Params = nil
	r.Id = nil
}

type serverResp struct {
	JsonRpc string          `json:"jsonrpc"`
	Result  interface{}     `json:"result"`
	Error   *errorResp      `json:"error,omitempty"`
	Id      json.RawMessage `json:"id,omitempty"`
}

type errorResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	c.req.reset()
	if err := c.dec.Decode(&c.req); err != nil {
		return err
	}
	r.ServiceMethod = c.req.Method

	c.mu.Lock()
	defer c.mu.Unlock()
	c.seq++
	c.pending[c.seq] = c.req.Id
	c.req.Id = nil
	r.Seq = c.seq
	return nil
}

func (c *serverCodec) ReadRequestBody(v interface{}) error {
	if v == nil {
		return nil
	}

}

var jsonNil = json.RawMessage("null")

func (c *serverCodec) WriteResponse(r *rpc.Response, v interface{}) error {
	c.mu.Lock()
	b, ok := c.pending[r.Seq]
	if !ok {
		c.mu.Unlock()
		return errors.New("invalid seq number in resp")
	}
	delete(c.pending, r.Seq)
	c.mu.Unlock()

	if b == nil {
		b = jsonNil
	}
	resp := serverResp{Id: b}
	if r.Error == "" {
		resp.Result = v
	} else {
		resp.Error = &errorResp{
			Code:    0,
			Message: r.Error,
			Data:    nil,
		}
	}
	return c.enc.Encode(resp)
}

func (c *serverCodec) Close() error {
	return c.closer.Close()
}

// ServeConn runs the JSON-RPC 2.0 server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
func ServeConn(conn io.ReadWriteCloser) {
	rpc.ServeCodec(NewServerCodec(conn))
}
