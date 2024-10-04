package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"nw-guardian/internal"
	"nw-guardian/internal/agent"
	"nw-guardian/internal/users"
	"time"
)

type Session struct {
	ID        primitive.ObjectID `json:"item_id" bson:"item_id"`
	IsAgent   bool               `json:"is_agent" bson:"is_agent"`
	SessionID primitive.ObjectID `json:"session_id" bson:"_id"`
	Expiry    time.Time          `json:"expiry" bson:"expiry"`
	Created   time.Time          `json:"created" bson:"created"`
	WSConn    string             `json:"ws_conn" bson:"ws_conn"`
	IP        string             `json:"ip,omitempty" bson:"ip"`
}

// Create a session from user id, and include expiry, return error if fails
func (s *Session) Create(db *mongo.Database) error {
	ee := internal.ErrorFormat{Package: "internal.auth", Level: log.ErrorLevel, Function: "session.Create"}

	s.SessionID = primitive.NewObjectID()
	s.Expiry = time.Now().Add(time.Hour * 24)
	s.Created = time.Now()

	if (s.ID == primitive.ObjectID{}) {
		ee.Message = "invalid id used to create session"
		return ee.ToError()
	}

	mar, err := bson.Marshal(s)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to marshal"
		return ee.ToError()
	}
	var b *bson.D
	err = bson.Unmarshal(mar, &b)
	if err != nil {
		ee.Message = "unable to unmarshal"
		return ee.ToError()
	}

	_, err = db.Collection("sessions").InsertOne(context.TODO(), b)
	if err != nil {
		ee.Message = "unable to insert session into db"
		ee.Error = err
		return ee.ToError()
	}

	return nil
}

// FromID returns a user if it finds a matching user with the provided ID
func (s *Session) FromID(db *mongo.Database) (*Session, error) {
	ee := internal.ErrorFormat{Package: "internal.auth", Level: log.ErrorLevel, Function: "session.FromID"}

	var filter = bson.D{{"_id", s.SessionID}}
	cursor, err := db.Collection("sessions").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		ee.Message = "unable to find sessions for id"
		ee.Error = err
		return nil, ee.ToError()
	}

	if len(results) < 1 {
		ee.Message = "no sessions found"
		ee.Error = err
		return nil, ee.ToError()
	}

	if len(results) > 1 {
		ee.Message = "multiple sessions found"
		ee.Error = err
		return nil, ee.ToError()
	}

	doc, err := bson.Marshal(&results[0])
	if err != nil {
		ee.Message = "unable to marshal first result"
		ee.Error = err
		return nil, ee.ToError()
	}

	var session *Session
	err = bson.Unmarshal(doc, &session)
	if err != nil {
		ee.Message = "unable to marshal session"
		ee.Error = err
		return nil, ee.ToError()
	}

	return session, nil
}

// GetAgent get the user from the token, otherwise return error
func GetAgent(token *jwt.Token, db *mongo.Database) (*agent.Agent, error) {
	ee := internal.ErrorFormat{Package: "internal.auth", Level: log.ErrorLevel, Function: "session.GetAgent"}

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
		ee.Message = "session is not an agent"
		return nil, ee.ToError()
	}

	if time.Now().After(s.Expiry) {
		ee.Error = errors.New("token expired")
		return nil, ee.ToError()
	}

	iId, err := primitive.ObjectIDFromHex(itemId)
	if err != nil {
		ee.Error = err
		return nil, ee.ToError()
	}

	if iId != s.ID {
		ee.Message = "id is invalid"
		return nil, ee.ToError()
	}

	a := &agent.Agent{ID: s.ID}
	err = a.Get(db)
	if err != nil {
		ee.Error = err
		return nil, ee.ToError()
	}

	return a, nil
}

func GetSessionID(token *jwt.Token) (primitive.ObjectID, error) {
	ee := internal.ErrorFormat{Package: "internal.auth", Level: log.ErrorLevel, Function: "session.GetSessionID"}

	claims := token.Claims.(jwt.MapClaims)
	sessionId := claims["session_id"].(string)
	sId, err := primitive.ObjectIDFromHex(sessionId)
	if err != nil {
		ee.Error = err
		ee.ObjectID = sId
		return primitive.NewObjectID(), ee.ToError()
	}

	return sId, nil
}

// GetUser get the user from the token, otherwise return error
func GetUser(token *jwt.Token, db *mongo.Database) (*users.User, error) {
	ee := internal.ErrorFormat{Package: "internal.auth", Level: log.ErrorLevel, Function: "session.GetUser"}

	claims := token.Claims.(jwt.MapClaims)
	itemId := claims["item_id"].(string)
	sessionId := claims["session_id"].(string)

	sId, err := primitive.ObjectIDFromHex(sessionId)
	if err != nil {
		ee.Error = err
		return nil, ee.ToError()
	}

	session := Session{SessionID: sId}
	s, err := session.FromID(db)
	if err != nil {
		ee.Error = err
		return nil, ee.ToError()
	}

	if time.Now().After(s.Expiry) {
		ee.Error = errors.New("token expired")
		return nil, ee.ToError()
	}

	iId, err := primitive.ObjectIDFromHex(itemId)
	if err != nil {
		ee.Error = err
		return nil, ee.ToError()
	}

	if iId != s.ID {
		ee.Error = errors.New("id mismatch")
		return nil, ee.ToError()
	}

	user := users.User{ID: s.ID}
	fromID, err := user.FromID(db)
	if err != nil {
		ee.Message = "unable to find user"
		ee.Error = err
		return nil, ee.ToError()
	}

	return fromID, nil
}

func (s *Session) UpdateConnWS(db *mongo.Database) error {
	ee := internal.ErrorFormat{Package: "internal.auth", Level: log.ErrorLevel, Function: "session.UpdateConnWS"}

	var filter = bson.D{{"_id", s.SessionID}}

	update := bson.D{{"$set", bson.D{{"ws_conn", s.WSConn}}}}

	_, err := db.Collection("sessions").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ee.Message = "unable to update ws connection session"
		ee.Error = err
		return ee.ToError()
	}
	log.Infof("Updated WSConn for Agent: %s WS: %s", s.ID, s.SessionID)

	return nil
}

func GetSessionFromWSConn(wsConn string, db *mongo.Database) (*Session, error) {
	ee := internal.ErrorFormat{Package: "internal.auth", Level: log.ErrorLevel, Function: "session.GetSessionFromWSConn"}

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
		ee.Error = errors.New("no session found")
		return nil, ee.ToError()
	}

	if len(results) > 1 {
		ee.Error = errors.New("multiple session found")
		return nil, ee.ToError()
	}

	doc, err := bson.Marshal(&results[0])
	if err != nil {
		ee.Message = "unable to marshal first result"
		ee.Error = err
		return nil, ee.ToError()
	}

	var session *Session
	err = bson.Unmarshal(doc, &session)
	if err != nil {
		ee.Message = "unable to unmarshal session"
		ee.Error = err
		return nil, ee.ToError()
	}

	return session, nil
}
