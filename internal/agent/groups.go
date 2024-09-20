package agent

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// todo add better error handling??!??

type Group struct {
	ID          primitive.ObjectID   `json:"id" bson:"_id"`
	SiteID      primitive.ObjectID   `json:"site,omitempty" bson:"site"`
	Agents      []primitive.ObjectID `json:"agents,omitempty" bson:"agents"`
	Name        string               `json:"name" bson:"name"`
	Description string               `bson:"description" json:"description"`
}

func (c *Group) Create(db *mongo.Database) error {
	c.ID = primitive.NewObjectID()

	mar, err := bson.Marshal(c)
	if err != nil {
		log.Errorf("error marshalling agent group when creating: %s", err)
		return err
	}

	var b *bson.D
	err = bson.Unmarshal(mar, &b)
	if err != nil {
		log.Errorf("error unmarhsalling agent group when creating: %s", err)
		return err
	}
	result, err := db.Collection("agent_groups").InsertOne(context.TODO(), b)
	if err != nil {
		log.Errorf("error inserting to database: %s", err)
		return err
	}

	fmt.Printf("created agent group with id: %v\n", result.InsertedID)
	return nil
}

// GetAll get all checks based on id, and &/or type
func (c *Group) GetAll(db *mongo.Database) ([]*Group, error) {
	var filter = bson.D{{"site", c.SiteID}}

	cursor, err := db.Collection("agent_groups").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	var agentCheck []*Group

	for _, rb := range results {
		m, err := bson.Marshal(&rb)
		if err != nil {
			log.Errorf("2 %s", err)
			return nil, err
		}
		var tC Group
		err = bson.Unmarshal(m, &tC)
		if err != nil {
			return nil, err
		}
		agentCheck = append(agentCheck, &tC)
	}
	return agentCheck, nil
}

/**

By default, agents are not apart of groups. These will likely be used to automagically select targets of agents.

Eg. if you have 2 groups, datacenters and customers, you could configure the ping tests to
connect to the group of "customer agents", from the datacenter agents, so when selecting.

When adding a probe, and you select one of the available groups as the destination/target,
the backend will automatically calculate the far end target IPs and return them to the agents on next request,
and that sort of thing. We will need to update the data structure to account for that though...
or do we just include the original target IP/target ID?? We need to somehow specify that the agent/probe is a, group probe
without making the probes impossible to update especially if you have a TON of the agents...

*/
