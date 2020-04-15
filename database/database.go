package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/tmazeika/testpass/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net"
	"net/url"
	"time"
)

const (
	dataDir = "data/"
	timeout = 10 * time.Second
)

var (
	ErrBadJson  = errors.New("bad json")
	ErrNotFound = errors.New("not found")
)

type Database struct {
	mgoDb       *mongo.Database
	attachments *mongo.Collection
	cases       *mongo.Collection
	suites      *mongo.Collection
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
		return nil, fmt.Errorf("connect DB: %v", err)
	}

	if err := client.Ping(newCtx(), readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping DB: %v", err)
	}

	mgoDb := client.Database("testpass")
	return &Database{
		mgoDb:       mgoDb,
		attachments: mgoDb.Collection("attachments"),
		cases:       mgoDb.Collection("cases"),
		suites:      mgoDb.Collection("suites"),
	}, nil
}

func (d *Database) Close() error {
	err := d.mgoDb.Client().Disconnect(newCtx())
	if err != nil {
		return fmt.Errorf("disconnect DB: %v", err)
	}
	return nil
}

func newCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	return ctx
}
