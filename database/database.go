package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/tmazeika/testpass/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net"
	"net/url"
	"time"
)

const (
	dataDir = "data/"
	timeout = 10 * time.Second
)

var (
	ErrInvalidModel = errors.New("invalid model")
	ErrNotFound     = errors.New("not found")

	validate = validator.New()
)

type WithContext struct {
	*Database
	ctx context.Context
}

func (d *WithContext) newContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(d.ctx, timeout)
}

func (d *WithContext) insert(collection *mongo.Collection, v interface{}) (string, error) {
	ctx, cancel := d.newContext()
	defer cancel()
	res, err := collection.InsertOne(ctx, v)
	if err != nil {
		return "", fmt.Errorf("insert: %v", err)
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("connect DB: %v", err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping DB: %v", err)
	}
	log.Printf("Connected to DB at %s\n", opts.GetURI())

	mgoDb := client.Database("testpass")
	return &Database{
		mgoDb:       mgoDb,
		attachments: mgoDb.Collection("attachments"),
		cases:       mgoDb.Collection("cases"),
		suites:      mgoDb.Collection("suites"),
	}, nil
}

func (d *Database) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := d.mgoDb.Client().Disconnect(ctx); err != nil {
		return fmt.Errorf("disconnect DB: %v", err)
	}
	return nil
}

func (d *Database) WithContext(ctx context.Context) *WithContext {
	return &WithContext{
		Database: d,
		ctx:      ctx,
	}
}
