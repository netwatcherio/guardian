package agent

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
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
	// pin will be used for "auth" as the password, the ID will stay the same
}

func (a *Agent) UpdateAgentDetails(db *mongo.Database, newName string, newLocation string) error {
	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{
		{"$set", bson.D{
			{"name", newName},
			{"location", newLocation},
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

func (a *Agent) UpdateTimestamp(db *mongo.Database) error {
	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{{"$set", bson.D{{"updatedAt", time.Now()}}}}

	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (a *Agent) Initialize(db *mongo.Database) error {
	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{{"$set", bson.D{{"initialized", true}}}}

	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAgent check based on provided agent ID in check struct
func DeleteAgent(db *mongo.Database, agentID primitive.ObjectID) error {
	// filter based on check ID
	var filter = bson.D{{"_id", agentID}}

	err := DeleteProbesByAgentID(db, agentID)
	if err != nil {
		return err
	}

	_, err = db.Collection("agents").DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
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
		return err
	}

	return nil
}
func (a *Agent) DeInitialize(db *mongo.Database) error {
	var filter = bson.D{{"_id", a.ID}}

	update := bson.D{{"$set", bson.D{{"initialized", false}}}}

	_, err := db.Collection("agents").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func GeneratePin(max int) string {
	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
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
		log.Errorf("error searching database for agent: %s", err)
		return err
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Errorf("error cursoring through agents: %s", err)
		return err
	}

	if len(results) > 1 {
		return errors.New("multiple agents match when getting using id") // edge case??
	}

	if len(results) == 0 {
		return errors.New("no agents match when getting using id")
	}

	doc, err := bson.Marshal(&results[0])
	if err != nil {
		log.Errorf("1 %s", err)
		return err
	}

	err = bson.Unmarshal(doc, &a)
	if err != nil {
		log.Errorf("2 %s", err)
		return err
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
		log.Errorf("error marshalling agent data when creating: %s", err)
		return err
	}
	var b *bson.D
	err = bson.Unmarshal(mar, &b)
	if err != nil {
		log.Errorf("error unmarhsalling agent data when creating: %s", err)
		return err
	}
	result, err := db.Collection("agents").InsertOne(context.TODO(), b)
	if err != nil {
		log.Errorf("error inserting to database: %s", err)
		return err
	}

	// also create netinfo probe
	probe := Probe{Agent: a.ID, Type: ProbeType_NETWORKINFO}
	err = probe.Create(db)
	if err != nil {
		return err
	}

	// also create system info probe
	ss := Probe{Agent: a.ID, Type: ProbeType_SYSTEMINFO}
	err = ss.Create(db)
	if err != nil {
		return err
	}

	fmt.Printf("created agent with id: %v\n", result.InsertedID)
	return nil
}
