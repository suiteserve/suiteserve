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

type NewSuiteRun struct {
	FailureTypes []SuiteFailureType `json:"failure_types,omitempty" bson:"failure_types,omitempty" validate:"dive"`
	Tags         []string           `json:"tags,omitempty" bson:",omitempty" validate:"unique,dive,required"`
	EnvVars      []SuiteEnvVar      `json:"env_vars,omitempty" bson:"env_vars,omitempty" validate:"dive"`
	Attachments  []string           `json:"attachments,omitempty" bson:",omitempty" validate:"dive,required"`
	PlannedCases uint               `json:"planned_cases" bson:"planned_cases"`
	CreatedAt    int64              `json:"created_at" bson:"created_at" validate:"gte=0"`
}

func (s *NewSuiteRun) CreatedAtTime() time.Time {
	return iToTime(s.CreatedAt)
}

type UpdateSuiteRun struct {
	Status     SuiteStatus `json:"status" validate:"oneof=created running finished"`
	FinishedAt int64       `json:"finished_at,omitempty" bson:"finished_at,omitempty" validate:"gte=0"`
}

func (s *UpdateSuiteRun) FinishedAtTime() time.Time {
	return iToTime(s.FinishedAt)
}

type SuiteRun struct {
	Id             interface{} `json:"id" bson:"_id,omitempty"`
	NewSuiteRun    `bson:",inline"`
	UpdateSuiteRun `bson:",inline"`
}

func (d *WithContext) NewSuiteRun(s NewSuiteRun) (string, error) {
	if err := validate.Struct(&s); err != nil {
		log.Printf("validate suite run: %v\n", err)
		return "", ErrInvalidModel
	}

	return d.insert(d.suites, SuiteRun{
		NewSuiteRun: s,
		UpdateSuiteRun: UpdateSuiteRun{
			Status: SuiteStatusCreated,
		},
	})
}

func (d *WithContext) UpdateSuiteRun(suiteId string, s UpdateSuiteRun) error {
	suiteOid, err := primitive.ObjectIDFromHex(suiteId)
	if err != nil {
		return fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	if err := validate.Struct(&s); err != nil {
		log.Printf("validate suite run: %v\n", err)
		return ErrInvalidModel
	}

	ctx, cancel := d.newContext()
	defer cancel()
	res, err := d.suites.UpdateOne(ctx, bson.M{
		"_id":   suiteOid,
	}, bson.M{
		"$set": &s,
	})
	if err != nil {
		return fmt.Errorf("update suite run: %v", err)
	} else if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *WithContext) SuiteRun(id string) (*SuiteRun, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	ctx, cancel := d.newContext()
	defer cancel()
	res := d.suites.FindOne(ctx, bson.M{"_id": oid})
	var suiteRun SuiteRun
	if err := res.Decode(&suiteRun); err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find suite run: %v", err)
	}
	return &suiteRun, nil
}

func (d *WithContext) AllSuiteRuns(sinceTime int64) ([]SuiteRun, error) {
	ctx, cancel := d.newContext()
	defer cancel()
	cursor, err := d.suites.Find(ctx, bson.M{
		"created_at": bson.M{"$gte": sinceTime},
	}, options.Find().SetSort(bson.D{
		{"created_at", 1},
		{"_id", 1},
	}))
	if err != nil {
		return nil, fmt.Errorf("find all suite runs: %v", err)
	}

	suiteRuns := make([]SuiteRun, 0)
	if err := cursor.All(ctx, &suiteRuns); err != nil {
		return nil, fmt.Errorf("decode all suite runs: %v", err)
	}
	return suiteRuns, nil
}

func (d *WithContext) DeleteSuiteRun(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	ctx, cancel := d.newContext()
	defer cancel()
	if err := d.DeleteAllCaseRuns(id); err != nil {
		return err
	}
	if _, err := d.suites.DeleteOne(ctx, bson.M{"_id": oid}); err != nil {
		return fmt.Errorf("delete suite run: %v", err)
	}
	return nil
}

func (d *WithContext) DeleteAllSuiteRuns() error {
	ctx, cancel := d.newContext()
	defer cancel()
	if _, err := d.cases.DeleteMany(ctx, bson.M{}); err != nil {
		return fmt.Errorf("delete all case runs before all suite runs: %v", err)
	}
	if _, err := d.suites.DeleteMany(ctx, bson.M{}); err != nil {
		return fmt.Errorf("delete all suite runs: %v", err)
	}
	return nil
}
