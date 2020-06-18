package repo

import "encoding/json"

type ChangeOp string

const (
	ChangeOpInsert ChangeOp = "insert"
	ChangeOpUpdate ChangeOp = "update"
)

type Change struct {
	Op   ChangeOp `json:"op"`
	Coll Coll     `json:"coll"`
}

type InsertChange struct {
	Change
	Doc json.RawMessage `json:"doc"`
}

func newInsertChange(coll Coll, doc json.RawMessage) *InsertChange {
	return &InsertChange{
		Change: Change{
			Op:   ChangeOpInsert,
			Coll: coll,
		},
		Doc: doc,
	}
}

type UpdateChange struct {
	Change
	Id      string                 `json:"id"`
	Updated map[string]interface{} `json:"updated,omitempty"`
	Deleted []string               `json:"deleted,omitempty"`
}

func newUpdateChange(coll Coll, id string, updated map[string]interface{}, deleted []string) *UpdateChange {
	return &UpdateChange{
		Change: Change{
			Op:   ChangeOpUpdate,
			Coll: coll,
		},
		Id:      id,
		Updated: updated,
		Deleted: deleted,
	}
}
