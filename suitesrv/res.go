package suitesrv

import (
	"fmt"
)

type response struct {
	Seq     int64                  `json:"seq"`
	Cmd     string                 `json:"cmd"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

func (r *response) Error() string {
	return fmt.Sprintf("[seq=%d] %s: %v", r.Seq, r.Cmd, r.Payload)
}

func newHelloResponse(seq int64) *response {
	return &response{
		Seq: seq,
		Cmd: "hello",
	}
}

func newCreatedResponse(seq int64, id string) *response {
	return &response{
		Seq: seq,
		Cmd: "created",
		Payload: map[string]interface{}{
			"id": id,
		},
	}
}

func newOkResponse(seq int64) *response {
	return &response{
		Seq: seq,
		Cmd: "ok",
	}
}
