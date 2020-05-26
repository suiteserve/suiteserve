package repo

import "encoding/json"

type ChangeOp string

const (
	ChangeOpInsert ChangeOp = "insert"
	ChangeOpUpdate ChangeOp = "update"
)

type Change struct {
	Op      ChangeOp    `json:"op"`
	Coll    Collection  `json:"coll"`
	Payload interface{} `json:"payload"`
}

func newChangeFromJson(op ChangeOp, coll Collection, payloadJson string) (*Change, error) {
	var payload interface{}
	if err := json.Unmarshal([]byte(payloadJson), &payload); err != nil {
		return nil, err
	}
	return &Change{op, coll, payload}, nil
}
