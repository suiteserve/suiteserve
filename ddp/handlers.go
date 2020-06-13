package ddp

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"time"
)

type Handler interface {
	Handle(io.ReadWriter) Handler
}

type handlerFunc func(io.ReadWriter) Handler

func (f handlerFunc) Handle(rw io.ReadWriter) Handler {
	return f(rw)
}

type Conn struct {
	methods map[string]reflect.Type
}

func NewConn() *Conn {
	return &Conn{
		methods: make(map[string]reflect.Type),
	}
}

func (c *Conn) RegisterMethod(name string, f interface{}) {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		panic("f is not a func")
	}
	if t.NumOut() > 2 {
		panic("f has more than two output parameters")
	}
	if t.NumOut() > 1 && !fHasErrOut(t) {
		panic("f has two output parameters but no error type")
	}
	c.methods[name] = t
}

func fHasErrOut(t reflect.Type) bool {
	for i := 0; i < t.NumOut(); i++ {
		if t.Out(i).AssignableTo(reflect.TypeOf((error)(nil))) {
			return true
		}
	}
	return false
}

func EstablishConnection() Handler {
	return reqRes(func(msgType string, msg json.RawMessage) (Handler, interface{}) {
		if msgType != "connect" {
			return EstablishConnection(), nil
		}
		var connectMsg connectMessage
		if err := json.Unmarshal(msg, &connectMsg); err != nil {
			return EstablishConnection(), newErrorMessage("invalid json connect obj", msg)
		}
		if connectMsg.Version != "1" {
			return nil, newFailedMessage("1")
		}
		return mainHandler(), newConnectedMessage(newSessionId())
	})
}

func mainHandler() Handler {
	return reqRes(func(msgType string, msg json.RawMessage) (Handler, interface{}) {
		switch msgType {
		case "ping":
			return mainHandler(), pingHandler(msg)
		}
		return nil, nil
	})
}

func pingHandler(msg json.RawMessage) interface{} {
	var pingMsg pingMessage
	if err := json.Unmarshal(msg, &pingMsg); err != nil {
		return newErrorMessage("invalid json ping obj", msg)
	}
	return newPongMessage(pingMsg.Id)
}

func (c *Conn) rpcHandler(msg json.RawMessage) interface{} {
	var methodMsg methodMessage
	if err := json.Unmarshal(msg, &methodMsg); err != nil {
		return newErrorMessage("invalid json method obj", msg)
	}
	method, ok := c.methods[methodMsg.Method]
	if !ok {
		return newErrorMessage(fmt.Sprintf("unknown method %q", methodMsg.Method), msg)
	}
	if len(methodMsg.Params) > method.NumIn() {
		return newErrorMessage("too many method params", msg)
	}
	in := make([]reflect.Value, len(methodMsg.Params))
	for i, p := range methodMsg.Params {
		val, _ := ToEjson(p)
		valType := reflect.TypeOf(val)
		inType := method.In(i)
		if valType.AssignableTo(inType) {
			in[i] = reflect.ValueOf(val)
		} else {
			return newErrorMessage(fmt.Sprintf(
				"got bad param type %q, want assignable to %q", valType, inType), msg)
		}
	}
	out := reflect.ValueOf(method).Call(in)
	fmt.Printf("%+v\n", out)
	return nil
}

var validClientMsgs = []string{
	"connect",
	"ping",
	"pong",
	"sub",
	"unsub",
	"method",
}

func reqRes(h func(msgType string, msg json.RawMessage) (Handler, interface{})) Handler {
	return handlerFunc(func(wr io.ReadWriter) Handler {
		dec := json.NewDecoder(wr)
		dec.UseNumber()
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			if err, ok := err.(net.Error); ok {
				log.Println(err)
				return nil
			}
			sendErrorMessage(wr, "invalid json", nil)
			return reqRes(h)
		}
		var msg message
		if err := json.Unmarshal(raw, &msg); err != nil {
			sendErrorMessage(wr, "invalid json obj", nil)
			return reqRes(h)
		}
		var valid bool
		for _, validMsg := range validClientMsgs {
			if validMsg == msg.Type {
				valid = true
				break
			}
		}
		if !valid {
			sendErrorMessage(wr, fmt.Sprintf("unknown msg type %q", msg.Type), raw)
			return reqRes(h)
		}
		next, res := h(msg.Type, raw)
		if res != nil {
			if err := json.NewEncoder(wr).Encode(&res); err != nil {
				if err, ok := err.(net.Error); ok {
					log.Println(err)
					return nil
				}
				panic(err)
			}
		}
		return next
	})
}

func newSessionId() string {
	b := make([]byte, 8)
	if _, err := rand.Reader.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x%x", time.Now().Unix(), b)
}
