package repo

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

type Collection string

const (
	AttachmentCollection Collection = "attachments"
	CaseCollection       Collection = "cases"
	LogCollection        Collection = "logs"
	SuiteCollection      Collection = "suites"
)

type Repos interface {
	Attachments(context.Context) AttachmentRepo
	Cases(context.Context) CaseRepo
	Changes() <-chan Change
	Logs(context.Context) LogRepo
	Suites(context.Context) SuiteRepo
	Close() error
}

type Entity struct {
	Id string `json:"id" bson:"_id,omitempty"`
}

type SoftDeleteEntity struct {
	Entity   `bson:",inline"`
	Deleted   bool  `json:"deleted"`
	DeletedAt int64 `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

func valuesToSlice(values []string, slice interface{}) error {
	v := "[" + strings.Join(values, ",") + "]"
	return json.Unmarshal([]byte(v), slice)
}

func nowTimeMillis() int64 {
	now := time.Now()
	// Doesn't use now.UnixNano() to avoid Y2K262.
	return now.Unix()*time.Second.Milliseconds() +
		time.Duration(now.Nanosecond()).Milliseconds()
}
