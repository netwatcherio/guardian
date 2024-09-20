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
	filter := bson.M{"_id": probeID}
	update := bson.M{
		"$set": bson.M{
			"config.target.0.target": newTarget,
		},
	}

	_, err := db.Collection("probes").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "UpdateProbeTarget", Level: log.ErrorLevel, ObjectID: probeID, Message: "unable to update probe target", Error: err}.ToError()
	}

	return nil
}

func (a *Agent) UpdateTimestamp(db *mongo.Database) error {
	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{{"$set", bson.D{{"updatedAt", time.Now()}}}}

	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	err = a.Get(db)
	if err != nil {
		return err
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
					return internal.ErrorFormat{Package: "internal.agent", Function: "UpdateTimestamp", Level: log.ErrorLevel, ObjectID: a.ID, Message: "unable to get version for agent", Error: err}.ToError()
				}

				splitVer = append(splitVer, atoi)
			}

			if splitVer[0] >= 1 && splitVer[1] >= 2 && splitVer[2] >= 1 {
				probe := Probe{Agent: a.ID}
				pps, err2 := probe.GetAllProbesForAgent(db)
				if err2 != nil {
					return internal.ErrorFormat{Package: "internal.agent", Function: "UpdateTimestamp", Level: log.ErrorLevel, ObjectID: a.ID, Message: "unable to get all probes for agent", Error: err2}.ToError()
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
						return internal.ErrorFormat{
							Package:  "agent",
							Function: "UpdateTimestamp",
							Level:    log.ErrorLevel,
							ObjectID: a.ID,
							Message:  "unable to create speedtest servers probe for agent",
							Error:    err}.ToError()
					}
				}

				if !hasSpeedtest {
					target := ProbeTarget{Target: "ok"}

					s3 := Probe{Agent: a.ID, Type: ProbeType_SPEEDTEST, Config: ProbeConfig{Target: []ProbeTarget{target}}}
					err = s3.Create(db)
					if err != nil {
						return internal.ErrorFormat{Package: "internal.agent", Function: "UpdateTimestamp", Level: log.ErrorLevel, ObjectID: a.ID, Message: "unable to create speedtest probe for agent", Error: err}.ToError()
					}
				}
			}
		}
	}

	return nil
}

func (a *Agent) UpdateAgentVersion(version string, db *mongo.Database) error {
	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{{"$set", bson.D{{"version", version}}}}

	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "UpdateAgentVersion", Level: log.ErrorLevel, ObjectID: a.ID, Message: "unable to update agent version", Error: err}.ToError()
	}

	return nil
}

func (a *Agent) Initialize(db *mongo.Database) error {
	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{{"$set", bson.D{{"initialized", true}}}}

	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Initialize", Level: log.ErrorLevel, ObjectID: a.ID, Message: "unable to initialize agent", Error: err}.ToError()
	}

	return nil
}

// DeleteAgent check based on provided agent ID in check struct
func DeleteAgent(db *mongo.Database, agentID primitive.ObjectID) error {
	// filter based on check ID
	var filter = bson.D{{"_id", agentID}}

	err := DeleteProbesByAgentID(db, agentID)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "DeleteAgent", Level: log.ErrorLevel, ObjectID: agentID, Message: "unable to delete probes by agent id", Error: err}.ToError()
	}

	_, err = db.Collection("agents").DeleteMany(context.TODO(), filter)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "DeleteAgent", Level: log.ErrorLevel, ObjectID: agentID, Message: "unable to delete agent by id", Error: err}.ToError()
	}

	return nil
}

func (a *Agent) Deactivate(db *mongo.Database) error {
	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{
		{"$set", bson.D{
			{"initialized", false},
			{"pin", GeneratePin(9)},
		}},
	}
	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Deactivate", Level: log.ErrorLevel, ObjectID: a.ID, Message: "unable to deactivate agent", Error: err}.ToError()
	}

	return nil
}
func (a *Agent) DeInitialize(db *mongo.Database) error {
	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{{"$set", bson.D{{"initialized", false}}}}

	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "DeInitialize", Level: log.ErrorLevel, ObjectID: a.ID, Message: "unable to de-initialize agent", Error: err}.ToError()
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
	var filter = bson.D{{"_id", a.ID}}

	cursor, err := db.Collection("agents").Find(context.TODO(), filter)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Get", Level: log.ErrorLevel, ObjectID: a.ID, Message: "unable to search for agent by id", Error: err}.ToError()
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Get", Level: log.ErrorLevel, ObjectID: a.ID, Message: "error cursoring through agents", Error: err}.ToError()
	}

	if len(results) > 1 {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Get", Level: log.ErrorLevel, ObjectID: a.ID, Message: "multiple agents match when getting using id", Error: err}.ToError() // edge case??
	}

	if len(results) == 0 {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Get", Level: log.ErrorLevel, ObjectID: a.ID, Message: "no agents match", Error: err}.ToError()
	}

	doc, err := bson.Marshal(&results[0])
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Get", Level: log.ErrorLevel, ObjectID: a.ID, Message: "unable to marshal get agents results[0]", Error: err}.ToError()
	}

	err = bson.Unmarshal(doc, &a)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Get", Level: log.ErrorLevel, ObjectID: a.ID, Message: "unable to marshal get agents result", Error: err}.ToError()
	}

	return nil
}

func (a *Agent) Create(db *mongo.Database) error {
	// todo handle to check if agent id is set and all that...
	a.Pin = GeneratePin(9)
	a.ID = primitive.NewObjectID()
	a.Initialized = false
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()

	mar, err := bson.Marshal(a)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Create", Level: log.ErrorLevel, ObjectID: a.ID, Message: "error marshalling agent data when creating agent", Error: err}.ToError()
	}
	var b *bson.D
	err = bson.Unmarshal(mar, &b)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Create", Level: log.ErrorLevel, ObjectID: a.ID, Message: "error unmarshalling agent when creating agent", Error: err}.ToError()
	}
	result, err := db.Collection("agents").InsertOne(context.TODO(), b)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Create", Level: log.ErrorLevel, ObjectID: a.ID, Message: "error creating agent", Error: err}.ToError()
	}

	// also create netinfo probe
	probe := Probe{Agent: a.ID, Type: ProbeType_NETWORKINFO}
	err = probe.Create(db)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Create", Level: log.ErrorLevel, ObjectID: a.ID, Message: "error creating network info probe", Error: err}.ToError()
	}

	// also create system info probe
	ss := Probe{Agent: a.ID, Type: ProbeType_SYSTEMINFO}
	err = ss.Create(db)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Create", Level: log.ErrorLevel, ObjectID: a.ID, Message: "error creating system info probe", Error: err}.ToError()
	}

	s2 := Probe{Agent: a.ID, Type: ProbeType_SPEEDTEST_SERVERS}
	err = s2.Create(db)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Create", Level: log.ErrorLevel, ObjectID: a.ID, Message: "error creating speedtest servers probe", Error: err}.ToError()
	}
	target := ProbeTarget{Target: "ok"}

	s3 := Probe{Agent: a.ID, Type: ProbeType_SPEEDTEST, Config: ProbeConfig{Target: []ProbeTarget{target}}}
	err = s3.Create(db)
	if err != nil {
		return internal.ErrorFormat{Package: "internal.agent", Function: "Create", Level: log.ErrorLevel, ObjectID: a.ID, Message: "error creating speedtest probe", Error: err}.ToError()
	}

	// todo output to loki??
	log.Info("created agent with id: %v\n", result.InsertedID)
	return nil
}
