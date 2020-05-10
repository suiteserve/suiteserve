package database

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type ChangeNamespace struct {
	Collection string `json:"collection" bson:"coll"`
}

type ChangeAttachmentDocument struct {
	Payload Attachment `json:"payload" bson:"fullDocument"`
}

type ChangeCaseDocument struct {
	Payload Case `json:"payload" bson:"fullDocument"`
}

type ChangeLogDocument struct {
	Payload LogMessage `json:"payload" bson:"fullDocument"`
}

type ChangeSuiteDocument struct {
	Payload Suite `json:"payload" bson:"fullDocument"`
}

type Change struct {
	Operation       string `json:"operation" bson:"operationType"`
	ChangeNamespace `bson:"ns"`
	Payload         interface{} `json:"payload" bson:"-"`
}

func (d *WithContext) Watch(fn func(Change)) error {
	ctx, cancel := d.newContext()
	defer cancel()
	res, err := d.mgoDb.Watch(ctx, bson.D{},
		options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		return fmt.Errorf("watch db: %v", err)
	}
	go func() {
		for res.Next(d.ctx) {
			if err := d.handleChange(fn, res); err != nil {
				log.Println(err)
			}
		}
		if res.Err() != nil {
			log.Printf("watch db: %v\n", res.Err())
		}
	}()
	return nil
}

func (d *WithContext) handleChange(fn func(Change), res *mongo.ChangeStream) error {
	var change Change
	if err := res.Decode(&change); err != nil {
		return fmt.Errorf("decode db change: %v", err)
	}

	switch change.Collection {
	case d.attachments.Name():
		var document ChangeAttachmentDocument
		if err := bson.Unmarshal(res.Current, &document); err != nil {
			return fmt.Errorf("decode db attachment change payload: %v", err)
		}
		change.Payload = document.Payload
	case d.cases.Name():
		var document ChangeCaseDocument
		if err := bson.Unmarshal(res.Current, &document); err != nil {
			return fmt.Errorf("decode db case change payload: %v", err)
		}
		change.Payload = document.Payload
	case d.logs.Name():
		var document ChangeLogDocument
		if err := bson.Unmarshal(res.Current, &document); err != nil {
			return fmt.Errorf("decode db log change payload: %v", err)
		}
		change.Payload = document.Payload
	case d.suites.Name():
		var document ChangeSuiteDocument
		if err := bson.Unmarshal(res.Current, &document); err != nil {
			return fmt.Errorf("decode db suite change payload: %v", err)
		}
		change.Payload = document.Payload
	default:
		return fmt.Errorf("unknown db change collection: %v", change.Collection)
	}

	go fn(change)
	return nil
}
