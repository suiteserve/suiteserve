package repo

import "encoding/json"

type ChangeOp string

const (
	ChangeOpInsert ChangeOp = "insert"
	ChangeOpUpdate ChangeOp = "update"
)

type Change interface {
	Operation() ChangeOp
	Collection() Coll
	DocId() string
}

type DocChange struct {
	Op   ChangeOp `json:"op"`
	Coll Coll     `json:"coll"`
	Id   string   `json:"id"`
}

func (c *DocChange) Operation() ChangeOp {
	return c.Op
}

func (c *DocChange) Collection() Coll {
	return c.Coll
}

func (c *DocChange) DocId() string {
	return c.Id
}

type InsertDocChange struct {
	DocChange
	Doc json.RawMessage `json:"doc"`
}

func newInsertDocChange(coll Coll, id string, doc json.RawMessage) *InsertDocChange {
	return &InsertDocChange{
		DocChange: DocChange{
			Op:   ChangeOpInsert,
			Coll: coll,
			Id:   id,
		},
		Doc: doc,
	}
}

type UpdateDocChange struct {
	DocChange
	Updated map[string]interface{} `json:"updated"`
	Deleted []string               `json:"deleted"`
}

func newUpdateDocChange(coll Coll, id string, updated map[string]interface{}, deleted []string) *UpdateDocChange {
	return &UpdateDocChange{
		DocChange: DocChange{
			Op:   ChangeOpUpdate,
			Coll: coll,
			Id:   id,
		},
		Updated: updated,
		Deleted: deleted,
	}
}
