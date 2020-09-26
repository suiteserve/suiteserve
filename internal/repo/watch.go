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
	w := Watcher{ch: make(chan Change)}
	s, err := r.db.Collection(coll).Watch(ctx, mongo.Pipeline{
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
		return &w
	}
	go func() {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			if cerr := s.Close(ctx); cerr != nil && w.err == nil {
				w.err = cerr
			}
		}()
		defer close(w.ch)
		for s.Next(ctx) {
			var e watchEvent
			if err := s.Decode(&e); err != nil {
				w.err = err
				return
			}
			b, err := json.Marshal(&e)
			if err != nil {
				panic(err)
			}
			w.ch <- Change{
				Msg:  b,
				Coll: e.coll,
			}
		}
		if s.Err() != nil && !errors.Is(s.Err(), context.Canceled) {
			w.err = s.Err()
		}
	}()
	return &w
}
