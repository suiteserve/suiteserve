package repo

type Changefeed []*Change

type ChangeOp string

const (
	ChangeOpInsert ChangeOp = "insert"
	ChangeOpUpdate ChangeOp = "update"
)

type Change struct {
	Id      string      `json:"id"`
	Op      ChangeOp    `json:"op"`
	Updated interface{} `json:"updated,omitempty"`
	Deleted interface{} `json:"deleted,omitempty"`
}
