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
	Coll Coll
	Msg  json.RawMessage
}

type watchEvent struct {
	Id     Id          `json:"id"`
	Insert interface{} `json:"insert,omitempty"`
	Update interface{} `json:"update,omitempty"`

	coll Coll
}

func (r *Repo) Watch(ctx context.Context) (<-chan Change, <-chan error) {
	changeCh := make(chan Change)
	errCh := make(chan error, 1)
	stream, err := r.db.Watch(ctx, mongo.Pipeline{
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
		}}},
		{{"$project", bson.D{
			{"id", 1},
			{"coll", 1},
			{"fullDocument", 1},
			{"update", 1},
		}}},
	})
	if err != nil {
		close(changeCh)
		errCh <- err
		return changeCh, errCh
	}
	go withErrChan(func() (err error) {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			safeClose(ctx, stream, &err)
		}()
		defer close(changeCh)
		for stream.Next(ctx) {
			var raw bson.Raw
			if err := stream.Decode(&raw); err != nil {
				return err
			}
			evt := bsonToWatchEvent(raw)
			changeCh <- Change{
				Msg:  mustMarshalJSON(&evt),
				Coll: evt.coll,
			}
		}
		if stream.Err() != nil && !errors.Is(stream.Err(), context.Canceled) {
			return stream.Err()
		}
		return nil
	}, errCh)
	return changeCh, errCh
}

type rawEvent struct {
	Id     Id
	Coll   Coll
	Insert bson.Raw `bson:"fullDocument"`
	Update bson.Raw
}

func bsonToWatchEvent(raw bson.Raw) watchEvent {
	var re rawEvent
	if err := bson.Unmarshal(raw, &re); err != nil {
		panic(err)
	}
	var as reflect.Type
	switch re.Coll {
	case Attachments:
		as = attachmentType
	case Suites:
		as = suiteType
	case Cases:
		as = caseType
	case Logs:
		as = logLineType
	default:
		panic(fmt.Sprintf("bad coll %q", re.Coll))
	}
	return watchEvent{
		Id:     re.Id,
		Insert: mustUnmarshalBSON(re.Insert, as),
		Update: mustUnmarshalBSON(re.Update, as),
		coll:   re.Coll,
	}
}

func mustUnmarshalBSON(raw bson.Raw, as reflect.Type) interface{} {
	if len(raw) == 0 {
		return nil
	}
	v := reflect.New(as).Interface()
	if err := bson.Unmarshal(raw, v); err != nil {
		panic(err)
	}
	return v
}
