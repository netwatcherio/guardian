package agent

import (
	"context"
	"crypto/rand"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"nw-guardian/internal"
	"regexp"
	"strconv"
	"time"
)

type Agent struct {
	ID               primitive.ObjectID `bson:"_id, omitempty"json:"id"`       // id
	Name             string             `bson:"name"json:"name"form:"name"`    // name of the agentprobe
	Site             primitive.ObjectID `bson:"site"json:"site"`               // _id of mongo object
	Pin              string             `bson:"pin"json:"pin"`                 // used for registration & authentication
	Initialized      bool               `bson:"initialized"json:"initialized"` // will this be used or will we use the sessions/jwt tokens?
	Location         string             `bson:"location"json:"location"`       // logical/physical location
	CreatedAt        time.Time          `bson:"createdAt"json:"createdAt"`
	UpdatedAt        time.Time          `bson:"updatedAt"json:"updatedAt"` // used for heart beat
	PublicIPOverride string             `bson:"public_ip_override"json:"public_ip_override"`
	Version          string             `bson:"version" json:"version"`
	// pin will be used for "auth" as the password, the ID will stay the same
}

func (a *Agent) UpdateAgentDetails(db *mongo.Database, newName string, newLocation string, newIP string) error {
	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{
		{"$set", bson.D{
			{"name", newName},
			{"location", newLocation},
			{"public_ip_override", newIP},
		}},
	}

	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	// Update the Agent struct to reflect the new state
	a.Name = newName
	a.Location = newLocation
	a.UpdatedAt = time.Now()

	return nil
}

func UpdateProbeTarget(db *mongo.Database, probeID primitive.ObjectID, newTarget string) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Function: "agent.UpdateProbeTarget", Level: log.ErrorLevel, ObjectID: probeID}

	filter := bson.M{"_id": probeID}
	update := bson.M{
		"$set": bson.M{
			"config.target.0.target": newTarget,
		},
	}

	_, err := db.Collection("probes").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ee.Message = "unable to update"
		ee.Error = err
		return ee.ToError()
	}

	return nil
}

func (a *Agent) UpdateTimestamp(db *mongo.Database) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Function: "agent.UpdateTimestamp", Level: log.ErrorLevel, ObjectID: a.ID}

	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{{"$set", bson.D{{"updatedAt", time.Now()}}}}

	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ee.Error = err
		return ee.ToError()
	}

	err = a.Get(db)
	if err != nil {
		ee.Error = err
		return ee.ToError()
	}

	pattern := `^v?(\d+)\.(\d+)\.(\d+)(rc|b|a)(\d+)$`
	re := regexp.MustCompile(pattern)

	versions := []string{a.Version}

	for _, version := range versions {
		versionMatch := re.FindStringSubmatch(version)
		if versionMatch != nil {
			/*fmt.Printf("Version: %s\n", version)
			fmt.Printf("  Major: %s\n", matches[1])
			fmt.Printf("  Minor: %s\n", matches[2])
			fmt.Printf("  Patch: %s\n", matches[3])
			fmt.Printf("  Release Type: %s\n", matches[4])
			fmt.Printf("  Release Number: %s\n", matches[5])
			fmt.Println()*/

			var splitVer []int
			for i, v := range versionMatch[1:] {
				if i == 3 {
					continue
				}
				atoi, err := strconv.Atoi(v)
				if err != nil {
					ee.Error = err
					ee.Message = "unable to get agent version"
					return ee.ToError()
				}

				splitVer = append(splitVer, atoi)
			}

			if splitVer[0] >= 1 && splitVer[1] >= 2 && splitVer[2] >= 1 {
				probe := Probe{Agent: a.ID}
				pps, err2 := probe.GetAllProbesForAgent(db)
				if err2 != nil {
					ee.Error = err2
					ee.Message = "unable to get all probes for agent"
					return ee.ToError()
				}

				hasSpeedtestServers := false
				hasSpeedtest := false

				for _, pp := range pps {
					if pp.Type == ProbeType_SPEEDTEST_SERVERS {
						hasSpeedtestServers = true
						continue
					} else if pp.Type == ProbeType_SPEEDTEST {
						hasSpeedtest = true
						continue
					}
				}

				if !hasSpeedtestServers {
					s2 := Probe{Agent: a.ID, Type: ProbeType_SPEEDTEST_SERVERS}
					err = s2.Create(db)
					if err != nil {
						ee.Error = err
						ee.Message = "unable to create speedtest servers probe for agent"
						return ee.ToError()
					}
				}

				if !hasSpeedtest {
					target := ProbeTarget{Target: "ok"}

					s3 := Probe{Agent: a.ID, Type: ProbeType_SPEEDTEST, Config: ProbeConfig{Target: []ProbeTarget{target}}}
					err = s3.Create(db)
					if err != nil {
						ee.Error = err
						ee.Message = "unable to create speedtest probe for agent"
						return ee.ToError()
					}
				}
			}
		}
	}

	return nil
}

func (a *Agent) UpdateAgentVersion(version string, db *mongo.Database) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Function: "agent.UpdateAgentVersion", Level: log.ErrorLevel, ObjectID: a.ID}

	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{{"$set", bson.D{{"version", version}}}}

	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to update agent version"
		return ee.ToError()
	}

	return nil
}

func (a *Agent) Initialize(db *mongo.Database) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Function: "agent.Initialize", Level: log.ErrorLevel, ObjectID: a.ID}

	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{{"$set", bson.D{{"initialized", true}}}}

	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ee.Message = "unable to initialize agent"
		ee.Error = err
		return ee.ToError()
	}

	return nil
}

// DeleteAgent check based on provided agent ID in check struct
func DeleteAgent(db *mongo.Database, agentID primitive.ObjectID) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Function: "agent.DeleteAgent", Level: log.ErrorLevel, ObjectID: agentID}

	// filter based on check ID
	var filter = bson.D{{"_id", agentID}}

	err := DeleteProbesByAgentID(db, agentID)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to delete probes by agent id"
		return ee.ToError()
	}

	_, err = db.Collection("agents").DeleteMany(context.TODO(), filter)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to delete agent by id"
		return ee.ToError()
	}

	return nil
}

func (a *Agent) Deactivate(db *mongo.Database) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Function: "agent.Deactivate", Level: log.ErrorLevel, ObjectID: a.ID}

	// todo should deactivating clear probe data??

	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{
		{"$set", bson.D{
			{"initialized", false},
			{"pin", GeneratePin(9)},
		}},
	}
	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to deactivate agent"
		return ee.ToError()
	}

	return nil
}
func (a *Agent) DeInitialize(db *mongo.Database) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Function: "agent.DeInitialize", Level: log.ErrorLevel, ObjectID: a.ID}

	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{{"$set", bson.D{{"initialized", false}}}}

	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ee.Message = "unable to de-initialize agent"
		ee.Error = err
		return ee.ToError()
	}

	return nil
}

func GeneratePin(max int) string {
	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		log.Error(err)
		return "6969420" // the gamer numbers (XD rawr)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

func (a *Agent) Get(db *mongo.Database) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Function: "agent.Get", Level: log.ErrorLevel, ObjectID: a.ID}

	var filter = bson.D{{"_id", a.ID}}

	cursor, err := db.Collection("agents").Find(context.TODO(), filter)
	if err != nil {
		ee.Message = "unable to search for agent by id"
		ee.Error = err
		return ee.ToError()
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		ee.Message = "error cursoring through agents"
		ee.Error = err
		return ee.ToError()
	}

	if len(results) > 1 {
		ee.Message = "too many results received"
		return ee.ToError()
	}

	if len(results) == 0 {
		ee.Message = "no agents found"
		return ee.ToError()
	}

	doc, err := bson.Marshal(&results[0])
	if err != nil {
		ee.Message = "unable to marshal results[0]"
		ee.Error = err
		return ee.ToError()
	}

	err = bson.Unmarshal(doc, &a)
	if err != nil {
		ee.Message = "unable to marshal agent"
		ee.Error = err
		return ee.ToError()
	}

	return nil
}

func (a *Agent) Create(db *mongo.Database) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Function: "agent.Create", Level: log.ErrorLevel, ObjectID: a.Site}

	// todo handle to check if agent id is set and all that...
	a.Pin = GeneratePin(9)
	a.ID = primitive.NewObjectID()
	a.Initialized = false
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()

	mar, err := bson.Marshal(a)
	if err != nil {
		ee.Message = "error marshalling agent creation"
		ee.Error = err
		return ee.ToError()
	}
	var b *bson.D
	err = bson.Unmarshal(mar, &b)
	if err != nil {
		ee.Message = "error unmarshalling agent creation"
		ee.Error = err
		return ee.ToError()
	}
	_, err = db.Collection("agents").InsertOne(context.TODO(), b)
	if err != nil {
		ee.Message = "error during agent creation"
		ee.Error = err
		return ee.ToError()
	}

	// also create netinfo probe
	probe := Probe{Agent: a.ID, Type: ProbeType_NETWORKINFO}
	err = probe.Create(db)
	if err != nil {
		ee.Message = "error creating netinfo probe"
		ee.Error = err
		return ee.ToError()
	}

	// also create system info probe
	ss := Probe{Agent: a.ID, Type: ProbeType_SYSTEMINFO}
	err = ss.Create(db)
	if err != nil {
		ee.Message = "error creating sysinfo probe"
		ee.Error = err
		return ee.ToError()
	}

	s2 := Probe{Agent: a.ID, Type: ProbeType_SPEEDTEST_SERVERS}
	err = s2.Create(db)
	if err != nil {
		ee.Message = "error creating speedtest servers probe"
		ee.Error = err
		return ee.ToError()
	}
	target := ProbeTarget{Target: "ok"}

	s3 := Probe{Agent: a.ID, Type: ProbeType_SPEEDTEST, Config: ProbeConfig{Target: []ProbeTarget{target}}}
	err = s3.Create(db)
	if err != nil {
		ee.Message = "error creating speedtest probe"
		ee.Error = err
		return ee.ToError()
	}

	// todo output to loki??
	//log.Infof("created agent with id: %s\n", result.InsertedID)
	return nil
}
