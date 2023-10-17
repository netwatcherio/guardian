package sites

import (
	"context"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SiteMemberRole string

const (
	SiteMemberRole_READONLY  SiteMemberRole = "READ_ONLY"  // view site only, no editing
	SiteMemberRole_READWRITE SiteMemberRole = "READ_WRITE" // view, add agents, probes, etc.
	SiteMemberRole_ADMIN     SiteMemberRole = "ADMIN"      // general admin, can add members, cannot remove owner or members
	SiteMemberRole_OWNER     SiteMemberRole = "OWNER"      // super admin
)

type SiteMember struct {
	User primitive.ObjectID `bson:"user"json:"user"`
	Role SiteMemberRole     `bson:"role"json:"role"`
	// roles: 0=READ ONLY, 1=READ-WRITE (Create only), 2=ADMIN (Delete Agents), 3=OWNER (Delete Sites)
	// ADMINS can regenerate agent pins
}

type NewSiteMember struct {
	Email string         `json:"email"form:"email"`
	Role  SiteMemberRole `json:"role"form:"role"`
}

// IsMember check if a user id is a member in the site
func (s *Site) IsMember(id primitive.ObjectID) bool {
	// check if the site contains the member with the provided id
	for _, k := range s.Members { // k is object id of member,
		if k.User == id {
			return true
		}
	}

	return false
}

// AddMember Add a member to the site then update document
func (s *Site) AddMember(id primitive.ObjectID, role SiteMemberRole, db *mongo.Database) error {
	// add member with the provided role
	if s.IsMember(id) {
		return errors.New("already a member")
	}

	newMember := SiteMember{
		User: id,
		Role: role,
	}

	s.Members = append(s.Members, newMember)
	j, _ := json.Marshal(s.Members)
	log.Warnf("%s", j)

	sites := db.Collection("sites")
	_, err := sites.UpdateOne(
		context.TODO(),
		bson.M{"_id": s.ID},
		bson.D{
			{"$set", bson.D{{"members", s.Members}}},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
