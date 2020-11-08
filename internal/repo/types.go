package repo

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"time"
)

type Id primitive.ObjectID

var nilId = Id(primitive.NilObjectID)

func NewId(hex string) (Id, error) {
	oid, err := primitive.ObjectIDFromHex(hex)
	return Id(oid), err
}

func (i Id) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(primitive.ObjectID(i))
}

func (i *Id) UnmarshalBSONValue(bt bsontype.Type, b []byte) error {
	var oid primitive.ObjectID
	err := bson.RawValue{Type: bt, Value: b}.Unmarshal(&oid)
	if err != nil {
		return err
	}
	*i = Id(oid)
	return nil
}

func (i Id) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

func (i *Id) UnmarshalJSON(b []byte) error {
	var hex string
	if err := json.Unmarshal(b, &hex); err != nil {
		return err
	}
	id, err := NewId(hex)
	if err != nil {
		return err
	}
	*i = id
	return nil
}

func (i Id) String() string {
	return primitive.ObjectID(i).Hex()
}

type MsTime time.Time

func NewMsTime(ms int64) MsTime {
	return MsTime(time.Unix(ms/1e3, (ms%1e3)*1e6))
}

func (t MsTime) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(time.Time(t))
}

func (t *MsTime) UnmarshalBSONValue(bt bsontype.Type, b []byte) error {
	var tt time.Time
	err := bson.RawValue{Type: bt, Value: b}.Unmarshal(&tt)
	if err != nil {
		return err
	}
	*t = MsTime(tt)
	return nil
}

func (t MsTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.toMs())
}

func (t *MsTime) UnmarshalJSON(b []byte) error {
	var ms int64
	if err := json.Unmarshal(b, &ms); err != nil {
		return err
	}
	*t = NewMsTime(ms)
	return nil
}

func (t MsTime) String() string {
	return strconv.FormatInt(t.toMs(), 10)
}

func (t MsTime) toMs() int64 {
	tt := time.Time(t)
	return tt.Unix()*1e3 + int64(tt.Nanosecond()/1e6)
}

func Bool(b bool) *bool {
	return &b
}

func Int64(i int64) *int64 {
	return &i
}

func String(s string) *string {
	return &s
}
