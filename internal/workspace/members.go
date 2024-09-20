package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Role string

const (
	MemberRole_READONLY  Role = "READ_ONLY"  // view site only, no editing
	MemberRole_READWRITE Role = "READ_WRITE" // view, add agents, probes, etc.
	MemberRole_ADMIN     Role = "ADMIN"      // general admin, can add members, cannot remove owner or members
	MemberRole_OWNER     Role = "OWNER"      // super admin
)

type Member struct {
	User primitive.ObjectID `bson:"user" json:"user"`
	Role Role               `bson:"role" json:"role"`
	// roles: 0=READ ONLY, 1=READ-WRITE (Create only), 2=ADMIN (Delete Agents), 3=OWNER (Delete Sites)
	// ADMINS can regenerate agent pins
}

type NewWorkspaceMember struct {
	Email string `json:"email" form:"email"`
	Role  Role   `json:"role" form:"role"`
}

func (s *Workspace) GetMemberRole(memberID primitive.ObjectID) (Role, error) {
	for _, member := range s.Members {
		if member.User == memberID {
			return member.Role, nil
		}
	}
	return "", fmt.Errorf("member with ID %s not found", memberID.Hex())
}

type MemberInfo struct {
	Email     string             `bson:"email"json:"email"` // email, will be used as username
	FirstName string             `bson:"firstName"json:"firstName"`
	LastName  string             `bson:"lastName"json:"lastName"`
	Role      Role               `bson:"role"json:"role"`
	ID        primitive.ObjectID `json:"id"bson:"_id"`
}

// UpdateMemberRole updates the role of a member in the site and the database
func (s *Workspace) UpdateMemberRole(memberID primitive.ObjectID, newRole Role, db *mongo.Database) error {
	// Find and update the member's role
	found := false
	for i, member := range s.Members {
		if member.User == memberID {
			s.Members[i].Role = newRole
			found = true
			break
		}
	}

	if !found {
		return errors.New("member not found")
	}

	// Update the site document in the database
	sites := db.Collection("sites")
	_, err := sites.UpdateOne(
		context.TODO(),
		bson.M{"_id": s.ID},
		bson.D{
			{"$set", bson.D{{"members", s.Members}}},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Workspace) GetMemberInfos(db *mongo.Database) ([]MemberInfo, error) {
	var memberInfos []MemberInfo

	for _, member := range s.Members {
		var memberInfo MemberInfo
		err := db.Collection("users").FindOne(context.TODO(), bson.M{"_id": member.User}).Decode(&memberInfo)
		if err != nil {
			// Handle the error, e.g., if the member is not found in the users collection
			log.Error(err)
			continue // or return nil, err if you prefer to stop on the first error
		}
		role, err := s.GetMemberRole(memberInfo.ID)
		if err != nil {
			log.Error(err)
		}
		memberInfo.Role = role
		memberInfos = append(memberInfos, memberInfo)
	}

	return memberInfos, nil
}

// IsMember check if a user id is a member in the site
func (s *Workspace) IsMember(id primitive.ObjectID) bool {
	// check if the site contains the member with the provided id
	for _, k := range s.Members { // k is object id of member,
		if k.User == id {
			return true
		}
	}

	return false
}

// AddMember Add a member to the site then update document
func (s *Workspace) AddMember(id primitive.ObjectID, role Role, db *mongo.Database) error {
	// add member with the provided role
	if s.IsMember(id) {
		return errors.New("already a member")
	}

	newMember := Member{
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

// RemoveMember removes a member from the site and updates the document
func (s *Workspace) RemoveMember(id primitive.ObjectID, db *mongo.Database) error {
	// Check if the member exists
	if !s.IsMember(id) {
		return errors.New("member not found")
	}

	// Remove the member with the provided ID
	var updatedMembers []Member
	for _, member := range s.Members {
		if member.User != id {
			updatedMembers = append(updatedMembers, member)
		}
	}
	s.Members = updatedMembers

	// Print the updated members
	j, _ := json.Marshal(s.Members)
	log.Warnf("%s", j)

	// Update the site document in the database
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
