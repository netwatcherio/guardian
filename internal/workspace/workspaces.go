package workspace

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"nw-guardian/internal/agent"
	"nw-guardian/internal/users"
	"time"
)

type Workspace struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `bson:"description" json:"description"`
	Location    string             `json:"location" bson:"location"` // logical/physical location
	Members     []Member           `json:"members" bson:"members"`
	// search for nested member id's when finding sites that belong to a user, is this more db intensive? does it matter? big O?
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

func (s *Workspace) UpdateSiteDetails(db *mongo.Database, newName, newLocation, newDescription string) error {
	filter := bson.D{{"_id", s.ID}}
	update := bson.D{
		{"$set", bson.D{
			{"name", newName},
			{"location", newLocation},
			{"description", newDescription},
			{"updatedAt", time.Now()},
		}},
	}

	_, err := db.Collection("sites").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	// Optionally, update the Workspace struct to reflect the new state
	s.Name = newName
	s.Location = newLocation
	s.Description = newDescription
	s.UpdatedAt = time.Now()

	return nil
}

func (s *Workspace) Create(owner primitive.ObjectID, db *mongo.Database) error {
	member := Member{
		User: owner,
		Role: MemberRole_OWNER,
	}

	s.Members = append(s.Members, member)
	s.ID = primitive.NewObjectID()
	s.CreatedAt = time.Now()

	mar, err := bson.Marshal(s)
	if err != nil {
		return errors.New("unable to marshal bson data")
	}
	var b *bson.D
	err = bson.Unmarshal(mar, &b)
	if err != nil {
		return errors.New("unable to marshal site data")
	}
	_, err = db.Collection("sites").InsertOne(context.TODO(), b)
	if err != nil {
		return errors.New("unable to create site")
	}

	// todo check this shit
	u := users.User{ID: member.User}
	usr, err := u.FromID(db)
	if err != nil {
		return errors.New("unable to get user from id")
	}
	u = *usr
	err = u.AddSite(s.ID, db)
	if err != nil {
		return errors.New("unable to add site to user")
	}

	return nil
}

// todo when deleting site remove from user document as well

func (s *Workspace) GetAgents(db *mongo.Database) ([]*agent.Agent, error) {
	var filter = bson.D{{"site", s.ID}}

	cursor, err := db.Collection("agents").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, errors.New("unable to search database for agents")
	}

	if len(results) == 0 {
		return nil, errors.New("no agents match when using id")
	}

	var agents []*agent.Agent
	for i := range results {
		doc, err := bson.Marshal(&results[i])
		if err != nil {
			return nil, err
		}
		var a *agent.Agent
		err = bson.Unmarshal(doc, &a)
		if err != nil {
			return nil, err
		}

		agents = append(agents, a)
	}

	return agents, nil
}

// AgentCount returns count of agents for a site, or an error if its not successful
func (s *Workspace) AgentCount(db *mongo.Database) (int, error) {
	var filter = bson.D{{"site", s.ID}}

	count, err := db.Collection("agents").CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func GetSitesForMember(memberID primitive.ObjectID, db *mongo.Database) ([]Workspace, error) {
	// Define a filter to match sites where at least one member has the specified user ID.
	filter := bson.M{"members": bson.M{"$elemMatch": bson.M{"user": memberID}}}

	// Find the matching sites in the "sites" collection.
	cursor, err := db.Collection("sites").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var matchingSites []Workspace

	// Iterate through the cursor and decode each site into a Workspace struct.
	for cursor.Next(context.Background()) {
		var site Workspace
		if err := cursor.Decode(&site); err != nil {
			return nil, err
		}
		matchingSites = append(matchingSites, site)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// Now 'matchingSites' contains an array of sites that have at least one member with the specified user ID.

	return matchingSites, nil
}

// Get a site from the provided ID
func (s *Workspace) Get(db *mongo.Database) error {

	var filter = bson.D{{"_id", s.ID}}

	cursor, err := db.Collection("sites").Find(context.TODO(), filter)
	if err != nil {
		return err
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		return err
	}

	if len(results) > 1 {
		return errors.New("multiple sites match when using id")
	}

	if len(results) == 0 {
		return errors.New("no sites match when using id")
	}

	doc, err := bson.Marshal(&results[0])
	if err != nil {
		return err
	}

	var site Workspace
	err = bson.Unmarshal(doc, &site)
	if err != nil {
		return err
	}

	s.Name = site.Name
	s.Description = site.Description
	s.Location = site.Location
	s.CreatedAt = site.CreatedAt
	s.UpdatedAt = site.UpdatedAt
	s.Members = site.Members

	return nil
}

// Delete data based on provided agent ID in checkData struct
func (s *Workspace) Delete(db *mongo.Database) error {
	// filter based on check ID
	var filter = bson.D{{"_id", s.ID}}

	_, err := db.Collection("sites").DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}

/*func (s *Workspace) SiteStats(db *mongo.Database) ([]*agent.Stats, error) {
	var agentStats []*agent.Stats

	agents, err := s.GetAgents(db)
	if err != nil {
		return nil, err
	}
	for _, a := range agents {
		stats, err := a.GetLatestStats(db)
		if err != nil {
			log.Error(err)
		}
		agentStats = append(agentStats, stats)
	}

	return agentStats, nil
}*/
