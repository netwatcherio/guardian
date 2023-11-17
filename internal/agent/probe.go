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

// todo update targets to be a struct instead of a simple string

// for group based target data, on  generation of the "targets" grabbed by the agent on connection
// it will grab the latest IPs of the agent and include those as the "target" it self to aide in automating
// ProbeTarget target string will automatically be populated if it is a group probe, if not, the normal target string will be used
type ProbeTarget struct {
	Target string             `json:"target,omitempty" bson:"target"`
	Agent  primitive.ObjectID `json:"agent,omitempty" bson:"agent"`
	Group  primitive.ObjectID `json:"group,omitempty" bson:"group"`
}

/*
when a list of probetargets is given, normal targets will only contain a target, and not an agent, etc
- this way we can then re-include the probetarget into the data it sends back to differentiate between targets
even though there is technically only 1 "probe"

*/

type ProbeConfig struct {
	Target   []ProbeTarget `json:"target" bson:"target"`
	Duration int           `json:"duration" bson:"duration"`
	Count    int           `json:"count" bson:"count"`
	Interval int           `json:"interval" bson:"interval"`
	Server   bool          `bson:"server" json:"server"`
	Pending  time.Time     `json:"pending" bson:"pending"` // timestamp of when it was made pending / invalidate it after 10 minutes or so?
}

type ProbeType string

const (
	ProbeType_RPERF       ProbeType = "RPERF"
	ProbeType_MTR         ProbeType = "MTR"
	ProbeType_PING        ProbeType = "PING"
	ProbeType_SPEEDTEST   ProbeType = "SPEEDTEST"
	ProbeType_NETWORKINFO ProbeType = "NETINFO"
)

type ProbeDataRequest struct {
	Limit          int64     `json:"limit"`
	StartTimestamp time.Time `json:"startTimestamp"`
	EndTimestamp   time.Time `json:"endTimestamp"`
	Recent         bool      `json:"recent"`
	Option         string    `json:"option"`
}

func (c *Probe) FindSimilarProbes(db *mongo.Database) ([]*Probe, error) {
	// todo finding similar probes is based on the targets and not groups currently

	if len(c.Config.Target) == 0 {
		return nil, errors.New("no targets defined in probe config")
	}

	var similarProbes []*Probe

	for _, target := range c.Config.Target {
		// Skip this target if it is part of a group.
		if (target.Group != primitive.ObjectID{0}) {
			continue
		}

		// Build the filter to find probes with the same target and agent.
		filter := bson.D{
			{"config.target", bson.D{
				{"$elemMatch", bson.D{
					{"target", target.Target},
					{"agent", primitive.ObjectID{0}},
					{"group", primitive.ObjectID{0}}, // Ensure the target is not part of a group.
				}},
			}},
		}

		// Query the database for probes with matching targets.
		cursor, err := db.Collection("probes").Find(context.TODO(), filter)
		if err != nil {
			return nil, err
		}

		var results []bson.D
		if err := cursor.All(context.TODO(), &results); err != nil {
			return nil, err
		}

		for _, r := range results {

			var pData Probe
			doc, err := bson.Marshal(r)
			if err != nil {
				log.Errorf("Error marshalling: %s", err)
				continue // Skip this result on error.
			}
			err = bson.Unmarshal(doc, &pData)
			if err != nil {
				log.Errorf("Error unmarshalling: %s", err)
				continue // Skip this result on error.
			}

			similarProbes = append(similarProbes, &pData)
		}
	}

	if len(similarProbes) <= 0 {
		return nil, errors.New("no similar probes found")
	}

	return similarProbes, nil
}

func (c *Probe) Create(db *mongo.Database) error {
	c.ID = primitive.NewObjectID()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

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
