package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"nw-guardian/internal"
	"nw-guardian/internal/agent"
	"nw-guardian/internal/users"
	"os"
	"time"
)

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login returns error on fail, nil on success
func (r *Login) Login(db *mongo.Database) (string, error) {
	if r.Email == "" {
		ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "auth.Login", Message: "invalid email address"}
		ee.Print()
		return "", fmt.Errorf(ee.Message)
	}

	u := users.User{Email: r.Email}
	user, err := u.FromEmail(db)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(r.Password))
	if err != nil {
		ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "auth.Login", ObjectID: user.ID, Message: "invalid password"}
		ee.Print()
		return "", fmt.Errorf(ee.Message)
	}

	session := Session{
		ID:      user.ID,
		IsAgent: false,
	}

	err = session.Create(db)
	if err != nil {
		ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "auth.Login", ObjectID: user.ID, Message: "unable to create session", Error: err}
		ee.Print()
		return "", fmt.Errorf(ee.Message)
	}

	// Create the Claims
	claims := jwt.MapClaims{
		"item_id":    session.ID.Hex(),
		"session_id": session.SessionID.Hex(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(os.Getenv("KEY")))
	if err != nil {
		ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "auth.Login", Message: "unable to generate session token", Error: err}
		ee.Print()
		return "", err
	}

	out := map[string]any{
		"token": t,
		"data":  *user,
	}

	bytes, err := json.Marshal(out)
	if err != nil {
		ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "auth.Login", Message: "unable to marshal token"}
		ee.Print()
		return "", fmt.Errorf(ee.Message)
	}

	return string(bytes), nil
}

type Register struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Company   string `json:"company"`
}

// Register returns error on fail, nil on success
func (r *Register) Register(db *mongo.Database) (string, error) {
	if r.FirstName == "" {
		return "", errors.New("invalid first name")
	}
	if r.LastName == "" {
		return "", errors.New("invalid last name")
	}
	if r.Email == "" {
		// TODO validate email
		return "", errors.New("invalid email name")
	}
	if r.Password == "" {
		return "", errors.New("invalid password, please ensure passwords match")
	}

	pwd, err := bcrypt.GenerateFromPassword([]byte(r.Password), 10)
	if err != nil {
		return "", err
	}

	user := users.User{
		Email:     r.Email,
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Company:   r.Company,
		Admin:     false,
		Password:  string(pwd),
		Verified:  false,
		CreatedAt: time.Now(),
	}

	err = user.Create(db)
	if err != nil {
		return "", err
	}

	session := Session{
		ID:      user.ID,
		IsAgent: false,
	}
	err = session.Create(db)
	if err != nil {
		return "", err
	}

	out, err := user.FromID(db)
	if err != nil {
		return "", err
	}
	// Create the Claims
	// todo make this more copiable? don't manually specify the mappedClaims?
	claims := jwt.MapClaims{
		"item_id":    session.ID.Hex(),
		"session_id": session.SessionID.Hex(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(os.Getenv("KEY")))
	if err != nil {
		return "", err
	}

	outMap := map[string]any{
		"token": t,
		"data":  *out,
	}

	bytes, err := json.Marshal(outMap)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

type AgentLogin struct {
	PIN          string `json:"pin"`
	ID           string `json:"id"`
	AgentVersion string `json:"version"`
}

// AgentLogin returns error on fail, nil on success
func (r *AgentLogin) AgentLogin(db *mongo.Database) (string, error) {
	if r.PIN == "" {
		return "", errors.New("invalid pin")
	}

	if r.ID == "" {
		return "", errors.New("invalid id")
	}

	aId, err := primitive.ObjectIDFromHex(r.ID)

	if err != nil {
		return "", err
	}

	u := agent.Agent{ID: aId}
	err = u.Get(db)
	if err != nil {
		return "", err
	}

	if u.Pin != r.PIN {
		return "", errors.New("pins do not match")
	}

	session := Session{
		ID:      u.ID,
		IsAgent: true,
	}

	err = session.Create(db)
	if err != nil {
		return "", err
	}

	err = u.UpdateAgentVersion(r.AgentVersion, db)
	if err != nil {
		log.Warnf("failed to update agent version for %s", u.ID)
	}

	// Create the Claims
	claims := jwt.MapClaims{
		"item_id":    session.ID.Hex(),
		"session_id": session.SessionID.Hex(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(os.Getenv("KEY")))
	if err != nil {
		return "", err
	}

	out := map[string]any{
		"token": t,
		"data":  u,
	}

	bytes, err := json.Marshal(out)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
