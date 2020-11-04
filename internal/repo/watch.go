package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type Change struct {
	Coll string
	Msg  json.RawMessage
}

type Watcher struct {
	ch  chan Change
	err error
}

func newWatcher() *Watcher {
	return &Watcher{ch: make(chan Change)}
}

func (w *Watcher) Ch() <-chan Change {
	return w.ch
}

func (w *Watcher) Err() error {
	return w.err
}

type watchEvent struct {
	Id     `json:"id"`
	Insert interface{}   `json:"insert,omitempty"`
	Update interface{}   `json:"update,omitempty"`
	Delete []interface{} `json:"delete,omitempty"`

	coll string
}

func (r *Repo) watch(ctx context.Context, coll string) *Watcher {
	w := newWatcher()
	stream, err := r.db.Collection(coll).Watch(ctx, mongo.Pipeline{
		{{"$match", bson.D{
			{"operationType", bson.D{
				{"$in", bson.A{
					"insert",
					"update",
				}},
			}},
		}}},
		{{"$set", bson.D{
			{"id", "$documentKey._id"},
			{"coll", "$ns.coll"},
			{"update", "$updateDescription.updatedFields"},
			{"delete", "$updateDescription.removedFields"},
		}}},
		{{"$project", bson.D{
			{"id", 1},
			{"coll", 1},
			{"fullDocument", 1},
			{"update", 1},
			{"delete", 1},
		}}},
	})
	if err != nil {
		w.err = err
		close(w.ch)
		return w
	}
	go func() {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			safeClose(ctx, stream, &w.err)
		}()
		defer close(w.ch)
		for stream.Next(ctx) {
			var raw bson.Raw
			if err := stream.Decode(&raw); err != nil {
				w.err = err
				return
			}
			evt := rawValueToWatchEvent(raw)
			w.ch <- Change{
				Msg:  mustMarshalJSON(&evt),
				Coll: evt.coll,
			}
		}
		if stream.Err() != nil && !errors.Is(stream.Err(), context.Canceled) {
			w.err = stream.Err()
		}
	}()
	return w
}

type rawEvent struct {
	Id     `bson:"id"`
	Coll   string        `bson:"coll"`
	Insert bson.Raw      `bson:"fullDocument"`
	Update bson.Raw      `bson:"update"`
	Delete []interface{} `bson:"delete"`
}

func rawValueToWatchEvent(raw bson.Raw) watchEvent {
	var re rawEvent
	if err := bson.UnmarshalWithRegistry(reg, raw, &re); err != nil {
		panic(err)
	}
	var as reflect.Type
	switch re.Coll {
	case attachmentsColl:
		as = attachmentType
	case suitesColl:
		as = suiteType
	case casesColl:
		as = caseType
	case logsColl:
		as = logLineType
	default:
		panic(fmt.Sprintf("unknown coll %q", re.Coll))
	}
	return watchEvent{
		Id:     re.Id,
		Insert: mustUnmarshalBSON(re.Insert, as),
		Update: mustUnmarshalBSON(re.Update, as),
		Delete: re.Delete,
		coll:   re.Coll,
	}
}

func mustUnmarshalBSON(raw bson.Raw, as reflect.Type) interface{} {
	if len(raw) == 0 {
		return nil
	}
	v := reflect.New(as).Interface()
	if err := bson.UnmarshalWithRegistry(reg, raw, v); err != nil {
		panic(err)
	}
	return v
}
