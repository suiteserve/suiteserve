package store

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const timeout = 10 * time.Second

type Store struct {
	db *mongo.Database
}

func New(host, user, pass string) (*Store, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	uri := fmt.Sprintf("mongodb://%s:%s@%s:27017", user, pass, host)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return &Store{
		db: client.Database("testpass"),
	}, nil
}
