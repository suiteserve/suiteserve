package repo

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const timeout = 10 * time.Second

type Entity struct {
	Id `json:"id" bson:"_id,omitempty"`
}

type VersionedEntity struct {
	Version *int64 `json:"version"`
}

type SoftDeleteEntity struct {
	Deleted   *bool `json:"deleted"`
	DeletedAt *Time `json:"deletedAt,omitempty" bson:"deleted_at,omitempty"`
}

type Repo struct {
	db *mongo.Database
}

var reg = bson.NewRegistryBuilder().
	RegisterTypeEncoder(idType, bsoncodec.ValueEncoderFunc(encodeIdValue)).
	RegisterTypeDecoder(idType, bsoncodec.ValueDecoderFunc(decodeIdValue)).
	Build()

func Open(addr, replSet, user, pass, db string) (*Repo, error) {
	opts := options.Client().
		SetHosts([]string{addr}).
		SetReplicaSet(replSet).
		SetAuth(options.Credential{
			AuthSource: db,
			Username:   user,
			Password:   pass,
		}).
		SetAppName("suiteserve").
		SetRegistry(reg)
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

func (r *Repo) insert(ctx context.Context, coll string,
	v interface{}) (Id, error) {
	res, err := r.db.Collection(coll).InsertOne(ctx, v)
	if err != nil {
		return nil, err
	}
	return res.InsertedID, nil
}

func (r *Repo) findByIdProj(ctx context.Context, coll string, id Id,
	proj, v interface{}) (interface{}, error) {
	opts := options.FindOne().SetProjection(proj)
	err := r.db.Collection(coll).FindOne(ctx, Entity{id}, opts).Decode(v)
	if err == mongo.ErrNoDocuments {
		return nil, errNotFound{}
	}
	return v, err
}

func (r *Repo) findById(ctx context.Context, coll string, id Id,
	v interface{}) (interface{}, error) {
	return r.findByIdProj(ctx, coll, id, nil, v)
}

func (r *Repo) updateById(ctx context.Context, coll string, id Id,
	set interface{}) error {
	res, err := r.db.Collection(coll).UpdateOne(ctx, Entity{id}, bson.D{
		{"$inc", bson.D{{"version", 1}}},
		{"$set", set},
	})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errNotFound{}
	}
	return nil
}

func (r *Repo) deleteById(ctx context.Context, coll string, id Id,
	at Time) error {
	return r.updateById(ctx, coll, id, bson.D{
		{"deleted", true},
		{"deleted_at", at},
	})
}

func readAll(ctx context.Context, v interface{},
	fn func() (*mongo.Cursor, error)) (interface{}, error) {
	c, err := fn()
	if err != nil {
		return nil, err
	}
	if err := c.All(ctx, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func readOne(ctx context.Context, v interface{},
	fn func() (*mongo.Cursor, error)) (res interface{}, err error) {
	c, err := fn()
	if err != nil {
		return nil, err
	}
	defer safeClose(ctx, c, &err)
	if !c.Next(ctx) {
		return v, nil
	}
	return v, c.Decode(v)
}

type ContextCloser interface {
	Close(ctx context.Context) error
}

func safeClose(ctx context.Context, c ContextCloser, err *error) {
	if cerr := c.Close(ctx); cerr != nil && *err == nil {
		*err = cerr
	}
}

func mustMarshalJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
