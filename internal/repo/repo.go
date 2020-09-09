package repo

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const timeout = 30 * time.Second

type Entity struct {
	Id string `json:"id"`
}

type VersionedEntity struct {
	Version int64 `json:"version"`
}

type SoftDeleteEntity struct {
	Deleted   bool  `json:"deleted"`
	DeletedAt int64 `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

type Repo struct {
	db *mongo.Database
}

func Open(addr, replSet, user, pass, db string) (*Repo, error) {
	opts := options.Client().
		SetHosts([]string{addr}).
		SetReplicaSet(replSet).
		SetAuth(options.Credential{
			AuthSource: db,
			Username:   user,
			Password:   pass,
		}).
		SetAppName("suiteserve")
	client, err := mongo.NewClient(opts)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	return &Repo{client.Database(db)}, nil
}

func (r *Repo) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return r.db.Client().Disconnect(ctx)
}

func mustMarshalJson(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
