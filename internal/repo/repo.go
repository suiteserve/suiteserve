package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"reflect"
	"time"
)

const timeout = 10 * time.Second

type Id interface{}

var idType = reflect.TypeOf((*Id)(nil)).Elem()

type Entity struct {
	Id `json:"id" bson:"_id,omitempty"`
}

type VersionedEntity struct {
	Version int64 `json:"version"`
}

type SoftDeleteEntity struct {
	Deleted   bool  `json:"deleted,omitempty"`
	DeletedAt int64 `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

type Repo struct {
	db *mongo.Database
}

func Open(addr, replSet, user, pass, db string) (*Repo, error) {
	reg := bson.NewRegistryBuilder().
		RegisterTypeDecoder(idType, bsoncodec.ValueDecoderFunc(decodeIdValue)).
		Build()
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

func decodeIdValue(_ bsoncodec.DecodeContext, vr bsonrw.ValueReader, v reflect.Value) error {
	if !v.IsValid() || v.Type() != idType {
		return bsoncodec.ValueDecoderError{
			Name:     "DecodeIdValue",
			Types:    []reflect.Type{idType},
			Received: v,
		}
	}
	if vr.Type() == bson.TypeNull {
		return vr.ReadNull()
	}
	oid, err := vr.ReadObjectID()
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(oid))
	return nil
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

func (r *Repo) findByIdProj(ctx context.Context, coll string, id Id, proj, v interface{}) error {
	filter := bson.D{{"_id", id}}
	opts := options.FindOne().SetProjection(proj)
	err := r.db.Collection(coll).FindOne(ctx, filter, opts).Decode(v)
	if err == mongo.ErrNoDocuments {
		return errNotFound{}
	}
	return err
}

func (r *Repo) findById(ctx context.Context, coll string, id Id, v interface{}) error {
	return r.findByIdProj(ctx, coll, id, nil, v)
}

func (r *Repo) updateById(ctx context.Context, coll string, id Id, set interface{}) error {
	res, err := r.db.Collection(coll).UpdateOne(ctx, bson.D{
		{"_id", id},
	}, bson.D{
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

func (r *Repo) deleteById(ctx context.Context, coll string, id Id, at int64) error {
	return r.updateById(ctx, coll, id, bson.D{
		{"deleted", true},
		{"deleted_at", at},
	})
}

func HexToId(hex string) (Id, error) {
	return primitive.ObjectIDFromHex(hex)
}
