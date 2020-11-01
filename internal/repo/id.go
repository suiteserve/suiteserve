package repo

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
)

type Id interface{}

var idType = reflect.TypeOf((*Id)(nil)).Elem()

type bsonId struct {
	Id
}

func (i bsonId) MarshalBSONValue() (bsontype.Type, []byte, error) {
	v := i.Id
	if s, ok := i.Id.(string); ok {
		var err error
		if v, err = primitive.ObjectIDFromHex(s); err != nil {
			return 0, nil, err
		}
	}
	return bson.MarshalValue(v)
}

func encodeIdValue(_ bsoncodec.EncodeContext, vw bsonrw.ValueWriter, v reflect.Value) error {
	if !v.IsValid() || v.Type() != idType {
		return bsoncodec.ValueEncoderError{
			Name:     "EncodeIdValue",
			Types:    []reflect.Type{idType},
			Received: v,
		}
	}
	switch x := v.Interface().(type) {
	case nil:
		return vw.WriteNull()
	case primitive.ObjectID:
		return vw.WriteObjectID(x)
	case string:
		oid, err := primitive.ObjectIDFromHex(x)
		if err != nil {
			return errBadFormat{fmt.Errorf("bad id: %v", err)}
		}
		return vw.WriteObjectID(oid)
	default:
		panic("exhausted allowed interface conversions for id type")
	}
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
