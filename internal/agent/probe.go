package agent

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Probe struct {
	Type          ProbeType          `json:"type"bson:"type"`
	ID            primitive.ObjectID `json:"id"bson:"_id"`
	Agent         primitive.ObjectID `json:"agent"bson:"agent"`
	CreatedAt     time.Time          `bson:"createdAt"json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt"json:"updatedAt"`
	Notifications bool               `json:"notifications"bson:"notifications"` // notifications will be emailed to anyone who has permissions on their account / associated with the site
	Config        ProbeConfig        `bson:"config"json:"config"`
}

type ProbeConfig struct {
	Target   string    `json:"target" bson:"target"`
	Duration int       `json:"duration" bson:"duration"`
	Count    int       `json:"count" bson:"count"`
	Interval int       `json:"interval" bson:"interval"`
	Server   bool      `bson:"server" json:"server"`
	Pending  time.Time `json:"pending" bson:"pending"` // timestamp of when it was made pending / invalidate it after 10 minutes or so?
}

type ProbeType string

const (
	ProbeType_RPERF       ProbeType = "RPERF"
	ProbeType_MTR         ProbeType = "MTR"
	ProbeType_PING        ProbeType = "PING"
	ProbeType_SPEEDTEST   ProbeType = "SPEEDTEST"
	ProbeType_NETWORKINFO ProbeType = "NETINFO"
)

type CheckRequest struct {
	Limit          int64     `json:"limit"`
	StartTimestamp time.Time `json:"start_timestamp"`
	EndTimestamp   time.Time `json:"end_timestamp"`
	Recent         bool      `json:"recent"`
}

func (c *Probe) Create(db *mongo.Database) error {
	c.ID = primitive.NewObjectID()

	mar, err := bson.Marshal(c)
	if err != nil {
		log.Errorf("error marshalling agent check when creating: %s", err)
		return err
	}

	var b *bson.D
	err = bson.Unmarshal(mar, &b)
	if err != nil {
		log.Errorf("error unmarhsalling agent check when creating: %s", err)
		return err
	}
	result, err := db.Collection("probes").InsertOne(context.TODO(), b)
	if err != nil {
		log.Errorf("error inserting to database: %s", err)
		return err
	}

	fmt.Printf("created agent check with id: %v\n", result.InsertedID)
	return nil
}

func (c *Probe) Get(db *mongo.Database) ([]*Probe, error) {
	var filter = bson.D{{"_id", c.ID}}

	if c.Agent != (primitive.ObjectID{0}) {
		filter = bson.D{{"agent", c.Agent}}
	}

	cursor, err := db.Collection("probes").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	fmt.Println(results)

	if c.Agent == (primitive.ObjectID{0}) {
		if len(results) > 1 {
			return nil, errors.New("multiple check match when using id")
		}

		if len(results) == 0 {
			return nil, errors.New("no sites match when using id")
		}

		doc, err := bson.Marshal(&results[0])
		if err != nil {
			log.Errorf("1 %s", err)
			return nil, err
		}

		err = bson.Unmarshal(doc, &c)
		if err != nil {
			log.Errorf("2 %s", err)
			return nil, err
		}

		return nil, nil
	} else {
		var agentChecks []*Probe

		for _, r := range results {
			var acData Probe
			doc, err := bson.Marshal(r)
			if err != nil {
				log.Errorf("1 %s", err)
				return nil, err
			}
			err = bson.Unmarshal(doc, &acData)
			if err != nil {
				log.Errorf("22 %s", err)
				return nil, err
			}

			agentChecks = append(agentChecks, &acData)
		}

		return agentChecks, nil
	}

	return nil, nil
}

// GetAll get all checks based on id, and &/or type
func (c *Probe) GetAll(db *mongo.Database) ([]*Probe, error) {
	var filter = bson.D{{"agent", c.Agent}}
	if c.Type != "" {
		filter = bson.D{{"agent", c.Agent}, {"type", c.Type}}
	}

	cursor, err := db.Collection("probes").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	var agentCheck []*Probe

	for _, rb := range results {
		m, err := bson.Marshal(&rb)
		if err != nil {
			log.Errorf("2 %s", err)
			return nil, err
		}
		var tC Probe
		err = bson.Unmarshal(m, &tC)
		if err != nil {
			return nil, err
		}
		agentCheck = append(agentCheck, &tC)
	}
	return agentCheck, nil
}

func (c *Probe) Update(db *mongo.Database) error {
	var filter = bson.D{{"_id", c.ID}}

	marshal, err := bson.Marshal(c)
	if err != nil {
		return err
	}

	var b bson.D
	err = bson.Unmarshal(marshal, &b)
	if err != nil {
		log.Errorf("error unmarhsalling agent data when creating: %s", err)
		return err
	}

	update := bson.D{{"$set", b}}

	_, err = db.Collection("probes").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// Delete check based on provided agent ID in check struct
func (c *Probe) Delete(db *mongo.Database) error {
	// filter based on check ID
	var filter = bson.D{{"_id", c.ID}}
	if (c.Agent != primitive.ObjectID{}) {
		filter = bson.D{{"agent", c.Agent}}
	}

	_, err := db.Collection("probes").DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}
