package agent

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	ProbeType_RPERF             ProbeType = "RPERF"
	ProbeType_MTR               ProbeType = "MTR"
	ProbeType_PING              ProbeType = "PING"
	ProbeType_SPEEDTEST         ProbeType = "SPEEDTEST"
	ProbeType_SPEEDTEST_SERVERS ProbeType = "SPEEDTEST_SERVERS"
	ProbeType_NETWORKINFO       ProbeType = "NETINFO"
	ProbeType_SYSTEMINFO        ProbeType = "SYSINFO"
	ProbeType_TRAFFICSIM        ProbeType = "TRAFFICSIM"
)

type ProbeDataRequest struct {
	Limit          int64     `json:"limit"`
	StartTimestamp time.Time `json:"startTimestamp"`
	EndTimestamp   time.Time `json:"endTimestamp"`
	Recent         bool      `json:"recent"`
	Option         string    `json:"option"`
}

func (probe *Probe) FindSimilarProbes(db *mongo.Database) ([]*Probe, error) {
	if len(probe.Config.Target) == 0 {
		return nil, errors.New("no targets defined in probe config")
	}

	// Remove type before getting probes
	probe.Type = ""

	allProbes, err := probe.Get(db)
	if err != nil {
		return nil, fmt.Errorf("error fetching probes: %w", err)
	}

	similarProbes := findSimilarProbes(allProbes, probe)

	if len(similarProbes) == 0 {
		return nil, errors.New("no similar probes found")
	}

	return similarProbes, nil
}

func findSimilarProbes(probes []*Probe, targetProbe *Probe) []*Probe {
	var similarProbes []*Probe

	for _, p := range probes {
		if len(p.Config.Target) == 0 {
			continue
		}

		for _, targetConfig := range targetProbe.Config.Target {
			if isSimilarProbe(p, targetConfig) {
				similarProbes = append(similarProbes, p)
				break // Move to the next probe once a match is found
			}
		}
	}

	return similarProbes
}

func isSimilarProbe(probe *Probe, targetConfig ProbeTarget) bool {
	for _, probeTarget := range probe.Config.Target {
		// Check for matching agent IDs
		if targetConfig.Agent != primitive.NilObjectID && targetConfig.Agent == probeTarget.Agent {
			return true
		}

		// Check for manual targets on the same agent
		if targetConfig.Agent == probeTarget.Agent &&
			targetConfig.Target != "" &&
			targetConfig.Target == probeTarget.Target {
			return true
		}

		// Check for matching group IDs (if implemented in the future)
		if targetConfig.Group != primitive.NilObjectID && targetConfig.Group == probeTarget.Group {
			return true
		}
	}
	return false
}

func (probe *Probe) Create(db *mongo.Database) error {
	probe.ID = primitive.NewObjectID()
	probe.CreatedAt = time.Now()
	probe.UpdatedAt = time.Now()

	mar, err := bson.Marshal(probe)
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

func (probe *Probe) Get(db *mongo.Database) ([]*Probe, error) {
	var filter = bson.D{{"_id", probe.ID}}

	if probe.Type != "" && probe.Agent != (primitive.ObjectID{0}) {
		filter = bson.D{{"agent", probe.Agent}, {"type", probe.Type}}
	} else if probe.Agent != (primitive.ObjectID{0}) {
		filter = bson.D{{"agent", probe.Agent}}
	}

	cursor, err := db.Collection("probes").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	//fmt.Println(results)
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
func (probe *Probe) GetAll(db *mongo.Database) ([]*Probe, error) {
	var filter = bson.D{{"agent", probe.Agent}}
	if probe.Type != "" {
		filter = bson.D{{"agent", probe.Agent}, {"type", probe.Type}}
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

func (probe *Probe) UpdateFirstProbeTarget(db *mongo.Database, targetStatus string) error {
	var filter = bson.D{{"_id", probe.ID}}

	get, err := probe.Get(db)
	if err != nil {
		return err
	}
	get[0].Config.Target[0].Target = targetStatus

	if get[0].Type == ProbeType_SPEEDTEST {
		get[0].Config.Pending = time.Now()
	}

	update := bson.D{
		{"$set", get[0]},
	}

	_, err = db.Collection("probes").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (probe *Probe) GetAllProbesForAgent(db *mongo.Database) ([]*Probe, error) {
	var filter = bson.D{{"agent", probe.Agent}}
	if probe.Type != "" {
		filter = bson.D{{"agent", probe.Agent}, {"type", probe.Type}}
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

		if len(tC.Config.Target) > 0 && !(tC.Config.Server && tC.Type == ProbeType_TRAFFICSIM) {
			if tC.Config.Target[0].Agent != (primitive.ObjectID{}) {
				// todo get the latest public ip of the agent, and use that as the target
				check := Probe{Agent: tC.Config.Target[0].Agent, Type: ProbeType_NETWORKINFO}

				// .Get will update it self instead of returning a list with a first object
				dd, err := check.Get(db)
				if err != nil {
					log.Error(err)
					continue
				}

				dd[0].Agent = primitive.ObjectID{0}
				data, err := dd[0].GetData(&ProbeDataRequest{Recent: true, Limit: 1}, db)
				if err != nil {
					log.Error(err)
					continue
				}

				a := Agent{ID: tC.Config.Target[0].Agent}
				err = a.Get(db)
				if err != nil {
					log.Error(err)
					continue
				}

				lastElement := data[len(data)-1]
				var netResult NetResult
				// todo this needs to be fixed for if the probe is a rperf probe,
				if a.PublicIPOverride != "" {
					netResult.PublicAddress = a.PublicIPOverride
				} else {
					switch v := lastElement.Data.(type) {
					case primitive.D:
						// Marshal primitive.D into BSON bytes
						bsonData, err := bson.Marshal(v)
						if err != nil {
							log.Error(err)
							continue
						}

						// Unmarshal BSON bytes into NetResult
						err = bson.Unmarshal(bsonData, &netResult)
						if err != nil {
							log.Error(err)
							continue
						}
					case primitive.M:
						// Data is in the form of primitive.M
						bsonData, err := bson.Marshal(v)
						if err != nil {
							log.Error(err)
							continue
						}
						err = bson.Unmarshal(bsonData, &netResult)
						if err != nil {
							log.Error(err)
							continue
						}
					default:
						log.Fatalf("Data is neither primitive.D nor primitive.M")
					}
				}

				if tC.Type == ProbeType_RPERF || tC.Type == ProbeType_TRAFFICSIM {
					// todo get rperf server based on the probe's agent ID, get the probe information for the "rperf server"
					// todo and use that as the target and account for the public ip or ip override
					// todo this is a bit of a hack, but it works for now
					var pp = Probe{Agent: tC.Config.Target[0].Agent, Type: tC.Type}
					agent, err := pp.GetAll(db)
					if err != nil {
						log.Error(err)
						continue
					}

					for _, probe := range agent {
						if probe.Config.Server && probe.Type == tC.Type {
							var port = strings.Split(probe.Config.Target[0].Target, ":")[1]
							tC.Config.Target[0].Target = netResult.PublicAddress + ":" + port
							break

						}
					}
				} else {
					tC.Config.Target[0].Target = netResult.PublicAddress
				}
			}
		} else if tC.Config.Server && tC.Type == ProbeType_TRAFFICSIM {
			clients, err := FindTrafficSimClients(db, tC.Agent)
			if err != nil {
				log.Error(err)
				continue
			}

			// ~clear the targets just to be sure~
			// turns out we don't want to do that because the first element
			// should actually be the binding ip / port of the server

			// tC.Config.Target = nil

			// we are rebuilding the target list
			for _, client := range clients {
				newTarget := ProbeTarget{Agent: client.Agent}
				tC.Config.Target = append(tC.Config.Target, newTarget)
			}
		}

		// append the target to the probe
		agentCheck = append(agentCheck, &tC)
	}

	// todo find all the probes that are trafficsim types for the other agents in the current workspace
	// and add them as targets to the server probe for the traffic sim if it's a traffic server sim

	return agentCheck, nil
}

func FindTrafficSimClients(db *mongo.Database, serverAgentID primitive.ObjectID) ([]*Probe, error) {
	// we shouldn't need to search based on all the sites / reduce them because we trust the backend / no one will
	// abuse the functionality of adding a traffic sim server to a site that it doesn't belong to / link to agent?

	// Assuming `serverAgentID` is the ID of the agent with the TRAFFICSIM server probe.

	// Step 1: Define the filter to find non-server TRAFFICSIM probes targeting this server agent.
	filter := bson.D{
		{"type", ProbeType_TRAFFICSIM},         // Filter for TRAFFICSIM type probes.
		{"config.server", false},               // Ensure these are not servers.
		{"config.target.agent", serverAgentID}, // Target must be the server agent.
	}

	// Step 2: Query the probes collection based on the defined filter.
	var clientProbes []*Probe
	cursor, err := db.Collection("probes").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	if err := cursor.All(context.TODO(), &clientProbes); err != nil {
		return nil, err
	}

	// `clientProbes` now contains all TRAFFICSIM client probes targeting the given server agent.
	return clientProbes, nil
}

func (probe *Probe) Update(db *mongo.Database) error {
	var filter = bson.D{{"_id", probe.ID}}

	marshal, err := bson.Marshal(probe)
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
func (probe *Probe) Delete(db *mongo.Database) error {
	// filter based on check ID
	var filter = bson.D{{"_id", probe.ID}}
	if (probe.Agent != primitive.ObjectID{}) {
		filter = bson.D{{"agent", probe.Agent}}
	}

	_, err := db.Collection("probes").DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}
