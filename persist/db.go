package persist

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const timeout = 10 * time.Second

type DB struct {
	mgo    *mongo.Database
	bucket *gridfs.Bucket
}

func New(host, user, pass string) (*DB, error) {
	ctx := newCtx()
	uri := fmt.Sprintf("mongodb://%s:%s@%s:27017", user, pass, host)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	mgo := client.Database("testpass")
	bucket, err := gridfs.NewBucket(mgo)
	if err != nil {
		return nil, err
	}

	return &DB{
		mgo:    mgo,
		bucket: bucket,
	}, nil
}

func (db *DB) Close() error {
	return db.mgo.Client().Disconnect(newCtx())
}

func newCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	return ctx
}
