package repo

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Watcher struct {
	ch  chan Change
	err error
}

func newWatcher() *Watcher {
	return &Watcher{ch: make(chan Change)}
}

type Change struct {
	Msg  json.RawMessage
	Coll string
}

func (w *Watcher) Ch() <-chan Change {
	return w.ch
}

func (w *Watcher) Err() error {
	return w.err
}

type watchEvent struct {
	Id     `json:"id" bson:"_id"`
	Insert interface{} `json:"insert,omitempty" bson:"fullDocument"`
	Update interface{} `json:"update,omitempty"`
	Delete interface{} `json:"delete,omitempty"`

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
			{"_id", "$documentKey._id"},
			{"coll", "$ns.coll"},
			{"update", "$updateDescription.updatedFields"},
			{"delete", "$updateDescription.removedFields"},
		}}},
		{{"$project", bson.D{
			{"fullDocument", 1},
			{"coll", 1},
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
			var evt watchEvent
			if err := stream.Decode(&evt); err != nil {
				w.err = err
				return
			}
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
