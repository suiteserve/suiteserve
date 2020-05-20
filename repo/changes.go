package repo

type ChangeOp string
type ChangeColl string

const (
	ChangeOpInsert ChangeOp = "insert"
	ChangeOpUpdate          = "update"

	ChangeCollAttachments ChangeColl = "attachments"
	ChangeCollCases                  = "cases"
	ChangeCollLogs                   = "logs"
	ChangeCollSuites                 = "suites"
)

type Change struct {
	Op      ChangeOp
	Coll    ChangeColl
	Payload interface{}
}
