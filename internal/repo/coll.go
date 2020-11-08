package repo

type Coll string

const (
	Attachments Coll = "attachments"
	Cases       Coll = "cases"
	Logs        Coll = "logs"
	Suites      Coll = "suites"

	attachments = string(Attachments)
	cases       = string(Cases)
	logs        = string(Logs)
	suites      = string(Suites)
)
