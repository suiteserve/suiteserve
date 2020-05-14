package database

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type SuiteStatus string

const (
	SuiteStatusRunning SuiteStatus = "running"
	SuiteStatusPassed              = "passed"
	SuiteStatusFailed              = "failed"
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
	Name         string             `json:"name,omitempty" bson:",omitempty"`
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
	Status     SuiteStatus `json:"status" validate:"oneof=running passed failed"`
	FinishedAt int64       `json:"finished_at,omitempty" bson:"finished_at,omitempty" validate:"gte=0"`
}

func (s *UpdateSuite) FinishedAtTime() time.Time {
	return iToTime(s.FinishedAt)
}

type Suite struct {
	Id          interface{} `json:"id" bson:"_id,omitempty"`
	NewSuite    `bson:",inline"`
	UpdateSuite `bson:",inline"`
	Deleted     bool  `json:"delete"`
	DeletedAt   int64 `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

type AllSuites struct {
	Running  uint64  `json:"running"`
	Finished uint64  `json:"finished"`
	More     bool    `json:"more"`
	Suites   []Suite `json:"suites,omitempty"`
}

func (s *Suite) DeletedAtTime() time.Time {
	return iToTime(s.DeletedAt)
}

func (d *WithContext) NewSuite(s NewSuite) (string, error) {
	if err := validate.Struct(&s); err != nil {
		log.Printf("validate suite: %v\n", err)
		return "", ErrInvalidModel
	}

	return d.insert(d.suites, Suite{
		NewSuite: s,
		UpdateSuite: UpdateSuite{
			Status: SuiteStatusRunning,
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
		"_id":     suiteOid,
		"deleted": false,
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

func (d *WithContext) AllSuites(afterId *string, limit *int64) (*AllSuites, error) {
	filter := bson.M{
		"deleted": false,
	}
	if afterId != nil {
		afterOid, err := primitive.ObjectIDFromHex(*afterId)
		if err != nil {
			return nil, fmt.Errorf("%w: parse object id", ErrNotFound)
		}
		filter["_id"] = bson.M{"$lt": afterOid}
	}
	suitesAgg := bson.A{
		bson.M{"$match": filter},
		bson.M{"$sort": bson.M{"_id": -1}},
	}
	if limit != nil {
		suitesAgg = append(suitesAgg, bson.M{"$limit": *limit})
	}
	ctx, cancel := d.newContext()
	defer cancel()
	cursor, err := d.suites.Aggregate(ctx, bson.A{
		bson.M{"$facet": bson.D{
			{"running", bson.A{
				bson.M{"$match": bson.M{
					"status":  SuiteStatusRunning,
					"deleted": false,
				}},
				bson.M{"$count": "count"},
			}},
			{"finished", bson.A{
				bson.M{"$match": bson.M{
					"status":  bson.M{"$ne": SuiteStatusRunning},
					"deleted": false,
				}},
				bson.M{"$count": "count"},
			}},
			{"more", bson.A{
				bson.M{"$sort": bson.M{
					"_id": 1,
				}},
				bson.M{"$limit": 1},
				bson.M{"$project": bson.M{
					"_id": 1,
				}},
			}},
			{"suites", suitesAgg},
		}},
		bson.M{"$set": bson.D{
			{"running", bson.M{
				"$arrayElemAt": bson.A{
					"$running.count", 0,
				},
			}},
			{"finished", bson.M{
				"$arrayElemAt": bson.A{
					"$finished.count", 0,
				},
			}},
			{"more", bson.M{
				"$ne": bson.A{
					bson.M{"$arrayElemAt": bson.A{
						"$suites._id", -1,
					}},
					bson.M{"$arrayElemAt": bson.A{
						"$more._id", 0,
					}},
				},
			}},
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("find all suites: %v", err)
	}

	suites := make([]AllSuites, 0)
	if err := cursor.All(ctx, &suites); err != nil {
		return nil, fmt.Errorf("decode all suites: %v", err)
	}
	return &suites[0], nil
}

func (d *WithContext) DeleteSuite(id string) (bool, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	ctx, cancel := d.newContext()
	defer cancel()
	if res, err := d.suites.UpdateOne(ctx, bson.M{
		"_id":     oid,
		"deleted": false,
	}, bson.M{
		"$set": bson.D{
			{"deleted", true},
			{"deleted_at", nowTimeMillis()},
		},
	}); err != nil {
		return false, fmt.Errorf("delete suite: %v", err)
	} else if res.MatchedCount == 0 {
		return false, nil
	}
	return true, nil
}

func (d *WithContext) DeleteAllSuites() error {
	ctx, cancel := d.newContext()
	defer cancel()
	if _, err := d.suites.UpdateMany(ctx,
		bson.M{
			"deleted": false,
		},
		bson.M{
			"$set": bson.D{
				{"deleted", true},
				{"deleted_at", nowTimeMillis()},
			},
		}); err != nil {
		return fmt.Errorf("delete all suites: %v", err)
	}
	return nil
}
