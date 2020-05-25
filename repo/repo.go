package repo

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"strings"
	"sync/atomic"
)

type Collection string

const (
	AttachmentCollection Collection = "attachments"
	CaseCollection       Collection = "cases"
	LogCollection        Collection = "logs"
	SuiteCollection      Collection = "suites"
)

type Repos interface {
	Attachments() AttachmentRepo
	Cases() CaseRepo
	Changes() <-chan Change
	Logs() LogRepo
	Suites() SuiteRepo
	StartedEmpty() bool
	Close() error
}

type Entity struct {
	Id string `json:"id" bson:"_id,omitempty"`
}

type SoftDeleteEntity struct {
	Entity    `bson:",inline"`
	Deleted   bool  `json:"deleted"`
	DeletedAt int64 `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

type IdGenerator func() string

var (
	IncIntIdGenerator = func() string {
		return strconv.FormatInt(atomic.AddInt64(&incIntId, 1), 10)
	}

	incIntId          int64 = 0
	uniqueIdGenerator       = func() string {
		return primitive.NewObjectID().Hex()
	}
)

func jsonValuesToArr(values []string, arr interface{}) error {
	v := "[" + strings.Join(values, ",") + "]"
	return json.Unmarshal([]byte(v), &arr)
}
