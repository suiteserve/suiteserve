package database

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type SuiteStatus string

const (
	SuiteStatusCreated  SuiteStatus = "created"
	SuiteStatusRunning              = "running"
	SuiteStatusFinished             = "finished"
)

type SuiteFailureType struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty" bson:",omitempty"`
}

type SuiteEnvVar struct {
	Key   string      `json:"key" validate:"required"`
	Value interface{} `json:"value,omitempty" bson:",omitempty"`
}

type NewSuite struct {
	FailureTypes []SuiteFailureType `json:"failure_types,omitempty" bson:"failure_types,omitempty" validate:"dive"`
	Tags         []string           `json:"tags,omitempty" bson:",omitempty" validate:"unique,dive,required"`
	EnvVars      []SuiteEnvVar      `json:"env_vars,omitempty" bson:"env_vars,omitempty" validate:"dive"`
	Attachments  []string           `json:"attachments,omitempty" bson:",omitempty" validate:"dive,required"`
	PlannedCases uint               `json:"planned_cases" bson:"planned_cases"`
	CreatedAt    int64              `json:"created_at" bson:"created_at" validate:"gte=0"`
}

func (s *NewSuite) CreatedAtTime() time.Time {
	return iToTime(s.CreatedAt)
}

type UpdateSuite struct {
	Status     SuiteStatus `json:"status" validate:"oneof=created running finished"`
	FinishedAt int64       `json:"finished_at,omitempty" bson:"finished_at,omitempty" validate:"gte=0"`
}

func (s *UpdateSuite) FinishedAtTime() time.Time {
	return iToTime(s.FinishedAt)
}

type Suite struct {
	Id          interface{} `json:"id" bson:"_id,omitempty"`
	NewSuite    `bson:",inline"`
	UpdateSuite `bson:",inline"`
}

func (d *WithContext) NewSuite(s NewSuite) (string, error) {
	if err := validate.Struct(&s); err != nil {
		log.Printf("validate suite: %v\n", err)
		return "", ErrInvalidModel
	}

	return d.insert(d.suites, Suite{
		NewSuite: s,
		UpdateSuite: UpdateSuite{
			Status: SuiteStatusCreated,
		},
	})
}

func (d *WithContext) UpdateSuite(suiteId string, s UpdateSuite) error {
	suiteOid, err := primitive.ObjectIDFromHex(suiteId)
	if err != nil {
		return fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	if err := validate.Struct(&s); err != nil {
		log.Printf("validate suite: %v\n", err)
		return ErrInvalidModel
	}

	ctx, cancel := d.newContext()
	defer cancel()
	res, err := d.suites.UpdateOne(ctx, bson.M{
		"_id": suiteOid,
	}, bson.M{
		"$set": &s,
	})
	if err != nil {
		return fmt.Errorf("update suite: %v", err)
	} else if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *WithContext) Suite(id string) (*Suite, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	ctx, cancel := d.newContext()
	defer cancel()
	res := d.suites.FindOne(ctx, bson.M{"_id": oid})
	var suite Suite
	if err := res.Decode(&suite); err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find and decode suite: %v", err)
	}
	return &suite, nil
}

func (d *WithContext) AllSuites(sinceTime int64) ([]Suite, error) {
	ctx, cancel := d.newContext()
	defer cancel()
	cursor, err := d.suites.Find(ctx, bson.M{
		"created_at": bson.M{"$gte": sinceTime},
	}, options.Find().SetSort(bson.D{
		{"created_at", 1},
		{"_id", 1},
	}))
	if err != nil {
		return nil, fmt.Errorf("find all suites: %v", err)
	}

	suites := make([]Suite, 0)
	if err := cursor.All(ctx, &suites); err != nil {
		return nil, fmt.Errorf("decode all suites: %v", err)
	}
	return suites, nil
}

func (d *WithContext) DeleteSuite(id string) (bool, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	ctx, cancel := d.newContext()
	defer cancel()
	//if err := d.DeleteAllCases(id); err != nil {
	//	return false, err
	//}
	if res, err := d.suites.DeleteOne(ctx, bson.M{"_id": oid}); err != nil {
		return false, fmt.Errorf("delete suite: %v", err)
	} else if res.DeletedCount == 0 {
		return false, nil
	}
	return true, nil
}

func (d *WithContext) DeleteAllSuites() error {
	ctx, cancel := d.newContext()
	defer cancel()
	// TODO: is it okay that transactions aren't being used?
	if err := d.suites.Drop(ctx); err != nil {
		return fmt.Errorf("delete all suites: %v", err)
	}
	if err := d.cases.Drop(ctx); err != nil {
		return fmt.Errorf("delete all cases: %v", err)
	}
	if err := d.logs.Drop(ctx); err != nil {
		return fmt.Errorf("delete all logs: %v", err)
	}
	return nil
}
