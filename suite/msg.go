package suite

import (
	"encoding/json"
	"fmt"
)

type msg struct {
	Seq     int64                  `json:"seq,omitempty"`
	Cmd     string                 `json:"cmd"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

func newMsg(msgJson interface{}) (*msg, error) {
	obj := msgJson.(map[string]interface{})
	seqJson, ok := obj["seq"].(json.Number)
	if !ok {
		return nil, errBadSeq(obj["seq"], "expected int")
	}
	seq, err := seqJson.Int64()
	if err != nil {
		return nil, errBadSeq(seqJson, "expected int")
	}
	cmd, ok := obj["cmd"].(string)
	if !ok {
		return nil, errBadCmd(seq, obj["cmd"], "expected string")
	}
	payload := make(map[string]interface{})
	payloadJson, ok := obj["payload"]
	if ok {
		if payload, ok = payloadJson.(map[string]interface{}); !ok {
			return nil, errBadPayload(seq, obj["payload"], "expected object")
		}
	}
	return &msg{
		Seq:     seq,
		Cmd:     cmd,
		Payload: payload,
	}, nil
}

func (m *msg) Error() string {
	return fmt.Sprintf("[seq %d] %s: %v", m.Seq, m.Cmd, m.Payload)
}
