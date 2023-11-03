package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"nw-guardian/internal/agent"
	"nw-guardian/internal/users"
	"time"
)

type Session struct {
	ID        primitive.ObjectID `json:"item_id"bson:"item_id"`
	IsAgent   bool               `json:"is_agent"bson:"is_agent"`
	SessionID primitive.ObjectID `json:"session_id"bson:"_id"`
	Expiry    time.Time          `json:"expiry"bson:"expiry"`
	WSConn    string             `json:"ws_conn"bson:"ws_conn"`
}

// Create a session from user id, and include expiry, return error if fails
func (s *Session) Create(db *mongo.Database) error {
	s.SessionID = primitive.NewObjectID()
	s.Expiry = time.Now().Add(time.Hour * 24)

	if (s.ID == primitive.ObjectID{}) {
		return errors.New("invalid item_id used to create session")
	}

	mar, err := bson.Marshal(s)
	if err != nil {
		return errors.New("something went wrong marshalling session struct")
	}
	var b *bson.D
	err = bson.Unmarshal(mar, &b)
	if err != nil {
		return errors.New("something went wrong marshalling session struct")
	}

	_, err = db.Collection("sessions").InsertOne(context.TODO(), b)
	if err != nil {
		return errors.New("something went wrong marshalling session struct")
	}

	return nil
}

// FromID returns a user if it finds a matching user with the provided ID
func (s *Session) FromID(db *mongo.Database) (*Session, error) {
	var filter = bson.D{{"_id", s.SessionID}}
	cursor, err := db.Collection("sessions").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	if len(results) < 1 {
		return nil, errors.New("no session found")
	}

	if len(results) > 1 {
		return nil, errors.New("multiple sessions found")
	}

	doc, err := bson.Marshal(&results[0])
	if err != nil {
		return nil, errors.New("something went wrong")
	}

	var session *Session
	err = bson.Unmarshal(doc, &session)
	if err != nil {
		log.Errorf("2 %s", err)
		return nil, errors.New("something went wrong unmarshalling session data")
	}

	return session, nil
}

// GetAgent get the user from the token, otherwise return error
func GetAgent(token *jwt.Token, db *mongo.Database) (*agent.Agent, error) {
	claims := token.Claims.(jwt.MapClaims)
	itemId := claims["item_id"].(string)
	sessionId := claims["session_id"].(string)

	sId, err := primitive.ObjectIDFromHex(sessionId)
	if err != nil {
		return nil, err
	}

	session := Session{SessionID: sId}
	s, err := session.FromID(db)
	if err != nil {
		return nil, err
	}

	if !s.IsAgent {
		return nil, errors.New("session is not valid agent session")
	}

	if time.Now().After(s.Expiry) {
		return nil, errors.New("token expired")
	}

	iId, err := primitive.ObjectIDFromHex(itemId)
	if err != nil {
		return nil, err
	}

	if iId != s.ID {
		return nil, errors.New("item id mismatch")
	}

	a := &agent.Agent{ID: s.ID}
	err = a.Get(db)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func GetSessionID(token *jwt.Token) (primitive.ObjectID, error) {
	claims := token.Claims.(jwt.MapClaims)
	sessionId := claims["session_id"].(string)
	sId, err := primitive.ObjectIDFromHex(sessionId)
	if err != nil {
		return primitive.NewObjectID(), err
	}

	return sId, nil
}

// GetUser get the user from the token, otherwise return error
func GetUser(token *jwt.Token, db *mongo.Database) (*users.User, error) {
	claims := token.Claims.(jwt.MapClaims)
	itemId := claims["item_id"].(string)
	sessionId := claims["session_id"].(string)

	sId, err := primitive.ObjectIDFromHex(sessionId)
	if err != nil {
		return nil, err
	}

	session := Session{SessionID: sId}
	s, err := session.FromID(db)
	if err != nil {
		return nil, err
	}

	if time.Now().After(s.Expiry) {
		return nil, errors.New("token expired")
	}

	iId, err := primitive.ObjectIDFromHex(itemId)
	if err != nil {
		return nil, err
	}

	if iId != s.ID {
		return nil, errors.New("item id mismatch")
	}

	user := users.User{ID: s.ID}
	fromID, err := user.FromID(db)
	if err != nil {
		return nil, err
	}

	return fromID, nil
}

func (s *Session) UpdateConnWS(db *mongo.Database) error {
	var filter = bson.D{{"_id", s.SessionID}}

	update := bson.D{{"$set", bson.D{{"ws_conn", s.WSConn}}}}

	_, err := db.Collection("sessions").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	log.Infof("Updated WSConn for Agent: %s WS: %s", s.ID, s.SessionID)

	return nil
}

func GetSessionFromWSConn(wsConn string, db *mongo.Database) (*Session, error) {
	var filter = bson.D{{"ws_conn", wsConn}}
	cursor, err := db.Collection("sessions").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	if len(results) < 1 {
		return nil, errors.New("no session found")
	}

	if len(results) > 1 {
		return nil, errors.New("multiple sessions found")
	}

	doc, err := bson.Marshal(&results[0])
	if err != nil {
		return nil, errors.New("something went wrong")
	}

	var session *Session
	err = bson.Unmarshal(doc, &session)
	if err != nil {
		log.Errorf("2 %s", err)
		return nil, errors.New("something went wrong unmarshalling session data")
	}

	return session, nil
}
