package agent

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"regexp"
	"strings"
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

// ProbeTarget for group based target data, on  generation of the "targets" grabbed by the agent on connection
// it will grab the latest IPs of the agent and include those as the "target" it self to aide in automating
// ProbeTarget target string will automatically be populated if it is a group probe, if not, the normal target string will be used
type ProbeTarget struct {
	Target string             `json:"target,omitempty" bson:"target"`
	Agent  primitive.ObjectID `json:"agent,omitempty" bson:"agent"`
	Group  primitive.ObjectID `json:"group,omitempty" bson:"group"`
}

type ProbeAlert struct {
	Agent     primitive.ObjectID `json:"agent,omitempty" bson:"agent" bson:"agent"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	Probe     Probe              `bson:"probe" json:"probe"`
	ProbeData ProbeData          `json:"probe_data" bson:"probeData"`
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

func DeleteProbesByAgentID(db *mongo.Database, agentID primitive.ObjectID) error {
	// todo if probe is deleted, delete associated data
	// todo if agent is delete, delete all probes, and data

	p := Probe{Agent: agentID}
	get, err := p.Get(db)
	if err != nil {
		return err
	}

	for _, probe := range get {
		err := DeleteProbeDataByProbeID(db, probe.ID)
		if err != nil {
			log.Error(err)
		}
	}

	// Convert the string ID to an ObjectID
	// Create a filter to match the document by ID
	filter := bson.M{"_id": agentID}

	// Perform the deletion
	_, err = db.Collection("probes").DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}

type ProbeType string

const (
	ProbeType_RPERF       ProbeType = "RPERF"
	ProbeType_MTR         ProbeType = "MTR"
	ProbeType_PING        ProbeType = "PING"
	ProbeType_SPEEDTEST   ProbeType = "SPEEDTEST"
	ProbeType_NETWORKINFO ProbeType = "NETINFO"
	ProbeType_SYSTEMINFO  ProbeType = "SYSINFO"
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

	/*get, err := c.Get(db)
	if err != nil {
		return nil, err
	}*/

	var similarProbes []*Probe

	for _, target := range c.Config.Target {
		// Skip this target if it is part of a group.
		if (target.Group != primitive.ObjectID{0}) {
			continue
		}

		var newT = strings.Split(target.Target, ":")
		var ttArget = target.Target
		if len(strings.Split(target.Target, ":")) >= 2 {
			ttArget = newT[0]
		}

		if (target.Group != primitive.ObjectID{0}) {
			continue
		}

		// Initialize the filter
		filter := bson.M{
			"config.target.agent": target.Agent,
		}

		// Check if an agent is defined and set the filter accordingly

		if (target.Agent != primitive.ObjectID{0}) {
			// If an agent is defined, use it in the filter instead of target host
			filter["config.target.agent"] = target.Agent
		} else {
			// If no agent is defined, build the filter to find probes with the same target host
			filter["config.target"] = bson.M{
				"$elemMatch": bson.M{
					"target": bson.M{"$regex": regexp.QuoteMeta(ttArget), "$options": "i"}, // The "i" option is for case-insensitive matching
					"agent":  primitive.ObjectID{},                                         // Assuming you want an empty ObjectID here
					"group":  primitive.ObjectID{},                                         // Ensure the target is not part of a group
				},
			}
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

	if c.Type != "" && c.Agent != (primitive.ObjectID{0}) {
		filter = bson.D{{"agent", c.Agent}, {"type", c.Type}}
	} else if c.Agent != (primitive.ObjectID{0}) {
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

func (c *Probe) GetAllProbesForAgent(db *mongo.Database) ([]*Probe, error) {
	var filter = bson.D{{"agent", c.Agent}}
	if c.Type != "" {
		filter = bson.D{{"agent", c.Agent}, {"type", c.Type}}
	}

	// this needs to be able to populate the target field with the ip/&port of the target based on
	// the public ip we grabbed from the agent previously, etc.

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

		if len(tC.Config.Target) > 0 {
			if tC.Config.Target[0].Agent != (primitive.ObjectID{}) {
				// todo get the latest public ip of the agent, and use that as the target
				check := Probe{Agent: tC.Config.Target[0].Agent, Type: ProbeType_NETWORKINFO}

				// .Get will update it self instead of returning a list with a first object
				dd, err := check.Get(db)
				if err != nil {
					return nil, err
				}

				dd[0].Agent = primitive.ObjectID{0}
				data, err := dd[0].GetData(&ProbeDataRequest{Recent: true, Limit: 1}, db)
				if err != nil {
					log.Warnf(err.Error())
					return nil, err
				}

				a := Agent{ID: tC.Config.Target[0].Agent}
				err = a.Get(db)
				if err != nil {
					return nil, err
				}

				lastElement := data[len(data)-1]
				var netResult NetResult
				if a.PublicIPOverride != "" {
					netResult.PublicAddress = a.PublicIPOverride
				} else {
					switch v := lastElement.Data.(type) {
					case primitive.D:
						// Marshal primitive.D into BSON bytes
						bsonData, err := bson.Marshal(v)
						if err != nil {
							log.Fatalf("Marshal failed: %v", err)
						}

						// Unmarshal BSON bytes into NetResult
						err = bson.Unmarshal(bsonData, &netResult)
						if err != nil {
							log.Fatalf("Unmarshal failed: %v", err)
						}
					case primitive.M:
						// Data is in the form of primitive.M
						bsonData, err := bson.Marshal(v)
						if err != nil {
							log.Fatalf("Marshal failed: %v", err)
						}
						err = bson.Unmarshal(bsonData, &netResult)
						if err != nil {
							log.Fatalf("Unmarshal failed: %v", err)
						}
					default:
						log.Fatalf("Data is neither primitive.D nor primitive.M")
					}
				}

				// todo this needs to be fixed for if the probe is a rperf probe,
				// todo because the target requires a port to be included

				tC.Config.Target[0].Target = netResult.PublicAddress
			}
		}

		// append the target to the probe
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
