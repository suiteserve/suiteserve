package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/tmazeika/testpass/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net"
	"net/url"
	"time"
)

const dbName = "testpass"
const timeout = 10 * time.Second

var (
	ErrNotFound = errors.New("entity not found")
)

type Database struct {
	mgoDb     *mongo.Database
	mgoBucket *gridfs.Bucket
}

func Open() (*Database, error) {
	host := config.Get(config.MongoHost, "localhost")
	port := config.Get(config.MongoPort, "27017")
	user := config.Get(config.MongoUser, "root")
	pass := config.Get(config.MongoPass, "pass")

	mongoUri := (&url.URL{
		Scheme: "mongodb",
		Host:   net.JoinHostPort(host, port),
	}).String()

	opts := options.Client()
	opts.SetAuth(options.Credential{
		Username: user,
		Password: pass,
	})
	opts.ApplyURI(mongoUri)

	client, err := mongo.Connect(newCtx(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect MongoDB: %v", err)
	}

	if err := client.Ping(newCtx(), readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	mgoDb := client.Database(dbName)
	mgoBucket, err := gridfs.NewBucket(mgoDb)
	if err != nil {
		return nil, fmt.Errorf("failed to create GridFS bucket: %v", err)
	}

	return &Database{mgoDb, mgoBucket}, nil
}

func (d *Database) Close() error {
	if err := d.mgoDb.Client().Disconnect(newCtx()); err != nil {
		return fmt.Errorf("failed to disconnect MongoDB: %v", err)
	}
	return nil
}

func newCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	return ctx
}
