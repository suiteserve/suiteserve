package repo

type ChangeOp string
type ChangeColl string

const (
	ChangeOpInsert ChangeOp = "insert"
	ChangeOpUpdate ChangeOp = "update"

	ChangeCollAttachments ChangeColl = "attachments"
	ChangeCollCases       ChangeColl = "cases"
	ChangeCollLogs        ChangeColl = "logs"
	ChangeCollSuites      ChangeColl = "suites"
)

type Change struct {
	Op      ChangeOp
	Coll    ChangeColl
	Payload interface{}
}
