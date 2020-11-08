package repo

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const timeout = 10 * time.Second

type Entity struct {
	Id *Id `json:"id" bson:"_id,omitempty"`
}

type VersionedEntity struct {
	Version *int64 `json:"version"`
}

type Repo struct {
	db *mongo.Database
}

// var reg = bson.NewRegistryBuilder().
// 	RegisterTypeEncoder(idType, bsoncodec.ValueEncoderFunc(encodeIdValue)).
// 	RegisterTypeDecoder(idType, bsoncodec.ValueDecoderFunc(decodeIdValue)).
// 	Build()

func Open(addr, replSet, user, pass, db string) (*Repo, error) {
	opts := options.Client().
		SetHosts([]string{addr}).
		SetReplicaSet(replSet).
		SetAuth(options.Credential{
			Username:   user,
			Password:   pass,
			AuthSource: db,
		}).
		SetAppName("suiteserve")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	return &Repo{db: client.Database(db)}, nil
}

func (r *Repo) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return r.db.Client().Disconnect(ctx)
}

func (r *Repo) insert(ctx context.Context, coll Coll,
	v interface{}) (Id, error) {
	res, err := r.db.Collection(string(coll)).InsertOne(ctx, v)
	if err != nil {
		return nilId, err
	}
	return Id(res.InsertedID.(primitive.ObjectID)), nil
}

func (r *Repo) findByIdProj(ctx context.Context, coll Coll, id Id,
	proj, v interface{}) error {
	filter := bson.D{{"_id", id}}
	opts := options.FindOne().SetProjection(proj)
	err := r.db.Collection(string(coll)).FindOne(ctx, filter, opts).Decode(v)
	if err == mongo.ErrNoDocuments {
		return errNotFound{}
	}
	return err
}

func (r *Repo) findById(ctx context.Context, coll Coll, id Id,
	v interface{}) error {
	return r.findByIdProj(ctx, coll, id, nil, v)
}

func (r *Repo) updateById(ctx context.Context, coll Coll, id Id,
	set interface{}) error {
	filter := bson.D{{"_id", id}}
	update := bson.D{
		{"$inc", bson.D{{"version", 1}}},
		{"$set", set},
	}
	res, err := r.db.Collection(string(coll)).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errNotFound{}
	}
	return nil
}

func readAll(ctx context.Context, v interface{},
	fn func() (*mongo.Cursor, error)) error {
	c, err := fn()
	if err != nil {
		return err
	}
	return c.All(ctx, v)
}

func readOne(ctx context.Context, v interface{},
	fn func() (*mongo.Cursor, error)) (err error) {
	c, err := fn()
	if err != nil {
		return err
	}
	defer safeClose(ctx, c, &err)
	if !c.Next(ctx) {
		return errNotFound{}
	}
	return c.Decode(v)
}

type ContextCloser interface {
	Close(ctx context.Context) error
}

func safeClose(ctx context.Context, closer ContextCloser, err *error) {
	if cerr := closer.Close(ctx); cerr != nil && *err == nil {
		*err = cerr
	}
}

func withErrChan(fn func() error, ch chan<- error) {
	ch <- fn()
}

func mustMarshalJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
