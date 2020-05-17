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
	"time"
)

const (
	storageDir = "storage/"
	timeout    = 10 * time.Second
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
	logs        *mongo.Collection
	suites      *mongo.Collection
}

func Open() (*Database, error) {
	host := config.Env(config.MongoHost, "localhost")
	port := config.Env(config.MongoPort, "27017")
	rs := config.Env(config.MongoReplicaSet, "rs0")
	user := config.Env(config.MongoUser, "testpass")
	pass := config.EnvFile(config.MongoPassFile, "testpass")

	opts := options.Client()
	opts.SetHosts([]string{net.JoinHostPort(host, port)})
	opts.SetReplicaSet(rs)
	opts.SetAuth(options.Credential{
		Username: user,
		Password: pass,
	})

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("connect db: %v", err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping db: %v", err)
	}
	log.Printf("Connected to DB at %v\n", opts.Hosts)

	mgoDb := client.Database("testpass")
	return &Database{
		mgoDb:       mgoDb,
		attachments: mgoDb.Collection("attachments"),
		cases:       mgoDb.Collection("cases"),
		logs:        mgoDb.Collection("logs"),
		suites:      mgoDb.Collection("suites"),
	}, nil
}

func (d *Database) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := d.mgoDb.Client().Disconnect(ctx); err != nil {
		return fmt.Errorf("disconnect db: %v", err)
	}
	return nil
}

func (d *Database) WithContext(ctx context.Context) *WithContext {
	return &WithContext{
		Database: d,
		ctx:      ctx,
	}
}

func iToTime(i int64) time.Time {
	if i < 0 {
		log.Printf("time i=%d must be non-negative\n", i)
		return time.Time{}
	}
	return time.Unix(i/1000, (i%1000)*time.Millisecond.Nanoseconds())
}

func nowTimeMillis() int64 {
	now := time.Now()
	// Doesn't use now.UnixNano() to avoid Y2K262.
	return now.Unix()*time.Second.Milliseconds() +
		time.Duration(now.Nanosecond()).Milliseconds()
}
