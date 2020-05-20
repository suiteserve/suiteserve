package repo

import (
	"encoding/json"
	"github.com/tidwall/buntdb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LogLevelType string

const (
	LogLevelTypeTrace LogLevelType = "trace"
	LogLevelTypeDebug              = "debug"
	LogLevelTypeInfo               = "info"
	LogLevelTypeWarn               = "warn"
	LogLevelTypeError              = "error"
)

type LogEntry struct {
	Id        interface{}  `json:"id" bson:"_id,omitempty"`
	Case      string       `json:"case"`
	Seq       int64        `json:"seq"`
	Level     LogLevelType `json:"level"`
	Trace     string       `json:"trace,omitempty" bson:",omitempty"`
	Message   string       `json:"message,omitempty" bson:",omitempty"`
	Timestamp int64        `json:"timestamp"`
}

type LogRepo interface {
	Save(LogEntry) (string, error)
	Find(id string) (*LogEntry, error)
	FindAllByCase(caseId string) ([]LogEntry, error)
}

type buntLogRepo struct {
	*buntRepo
}

func (r *buntRepo) newLogRepo() (*buntLogRepo, error) {
	err := r.db.ReplaceIndex("logs_case_id", "logs:*",
		buntdb.IndexJSON("case"), buntdb.IndexJSON("id"))
	if err != nil {
		return nil, err
	}
	return &buntLogRepo{r}, nil
}

func (r *buntLogRepo) Save(e LogEntry) (string, error) {
	b, err := json.Marshal(&e)
	if err != nil {
		return "", err
	}
	var id string
	err = r.db.Update(func(tx *buntdb.Tx) error {
		id = primitive.NewObjectID().Hex()
		e.Id = id
		_, _, err = tx.Set("logs:"+id, string(b), nil)
		if err != nil {
			return err
		}
		r.changes <- Change{
			Op:      ChangeOpInsert,
			Coll:    ChangeCollLogs,
			Payload: e,
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *buntLogRepo) Find(id string) (*LogEntry, error) {
	var e LogEntry
	err := r.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get("logs:" + id)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(v), &e)
	})
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *buntLogRepo) FindAllByCase(caseId string) ([]LogEntry, error) {
	values := make([]string, 0)
	err := r.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual("logs_case_id", caseId, func(k, v string) bool {
			values = append(values, v)
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	entries := make([]LogEntry, len(values))
	for i, v := range values {
		var e LogEntry
		if err := json.Unmarshal([]byte(v), &e); err != nil {
			return nil, err
		}
		entries[i] = e
	}
	return entries, nil
}
