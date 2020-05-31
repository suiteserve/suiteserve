package suite

import (
	"fmt"
)

type msg struct {
	Seq     int64                  `json:"seq,omitempty"`
	Cmd     string                 `json:"cmd"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

func (m *msg) Error() string {
	return fmt.Sprintf("[seq %d] %s: %v", m.Seq, m.Cmd, m.Payload)
}
