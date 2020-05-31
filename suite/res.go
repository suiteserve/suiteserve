package suite

import (
	"fmt"
)

type responseType string

const (
	typeSetSuiteId responseType = "set_suite_id"
)

type response struct {
	Type    responseType `json:"type"`
	Seq     int64        `json:"seq,omitempty"`
	Payload interface{}  `json:"payload,omitempty"`
}

func (r *response) Error() string {
	return fmt.Sprintf("[seq %d] %s: %v", r.Seq, r.Type, r.Payload)
}

func newSetSuiteIdResponse(seq int64, id string) *response {
	return &response{
		Type: typeSetSuiteId,
		Seq:  seq,
		Payload: map[string]interface{}{
			"id": id,
		},
	}
}
