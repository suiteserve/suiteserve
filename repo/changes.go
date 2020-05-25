package repo

type ChangeOp string

const (
	ChangeOpInsert ChangeOp = "insert"
	ChangeOpUpdate ChangeOp = "update"
)

type Change struct {
	Op      ChangeOp
	Coll    Collection
	Payload interface{}
}
