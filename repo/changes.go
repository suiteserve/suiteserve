package repo

type ChangeOp string

const (
	ChangeOpInsert ChangeOp = "insert"
	ChangeOpUpdate ChangeOp = "update"
)

type Change struct {
	Op         ChangeOp
	Collection Collection
	Payload    interface{}
}
