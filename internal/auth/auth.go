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
func (r *Login) Login(ip string, db *mongo.Database) (string, error) {
	ee := internal.ErrorFormat{Package: "internal.auth", Level: log.ErrorLevel, Function: "auth.Login"}
	if r.Email == "" {
		ee.Message = "invalid email"
		ee.Print()
		return "", fmt.Errorf(ee.Message)
	}

	u := users.User{Email: r.Email}
	user, err := u.FromEmail(db)
	if err != nil {
		ee.Message = "invalid email"
		ee.Print()
		return "", fmt.Errorf(ee.Message)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(r.Password))
	if err != nil {
		ee.Message = "invalid password"
		ee.Message += " - connecting ip: " + ip
		ee.Print()
		return "", fmt.Errorf(ee.Message)
	}

	session := Session{
		ID:      user.ID,
		IsAgent: false,
		IP:      ip,
	}

	err = session.Create(db)
	if err != nil {
		ee.Message = "unable to create session"
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
		ee.Message = "unable to sign token string"
		ee.Print()
		return "", fmt.Errorf(ee.Message)
	}

	out := map[string]any{
		"token": t,
		"data":  *user,
	}

	bytes, err := json.Marshal(out)
	if err != nil {
		ee.Message = "unable to marshal"
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
	ee := internal.ErrorFormat{Package: "internal.auth", Level: log.ErrorLevel, Function: "auth.Register"}

	if r.FirstName == "" {
		ee.Message = "invalid first name"
		ee.Print()
		return "", errors.New(ee.Message)
	}
	if r.LastName == "" {
		ee.Message = "invalid last name"
		ee.Print()
		return "", errors.New(ee.Message)
	}
	if r.Email == "" {
		ee.Message = "invalid email"
		ee.Print()
		// TODO validate email
		return "", errors.New(ee.Message)
	}
	if r.Password == "" {
		ee.Message = "invalid password"
		ee.Print()
		return "", errors.New(ee.Message)
	}

	pwd, err := bcrypt.GenerateFromPassword([]byte(r.Password), 10)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to generate bcrypt password"
		return "", ee.ToError()
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
		ee.Error = err
		ee.Message = "unable to create user"
		return "", ee.ToError()
	}

	session := Session{
		ID:      user.ID,
		IsAgent: false,
	}
	err = session.Create(db)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to create session"
		return "", ee.ToError()
	}

	out, err := user.FromID(db)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to get user from id"
		return "", ee.ToError()
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
		ee.Error = err
		ee.Message = "unable to create token signed string"
		return "", ee.ToError()
	}

	outMap := map[string]any{
		"token": t,
		"data":  *out,
	}

	bytes, err := json.Marshal(outMap)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to marshal session output"
		return "", ee.ToError()
	}

	return string(bytes), nil
}

type AgentLogin struct {
	PIN          string `json:"pin"`
	ID           string `json:"id"`
	AgentVersion string `json:"version"`
}

// AgentLogin returns error on fail, nil on success
func (r *AgentLogin) AgentLogin(ip string, db *mongo.Database) (string, error) {
	ee := internal.ErrorFormat{Package: "internal.auth", Level: log.ErrorLevel, Function: "auth.AgentLogin"}

	if r.PIN == "" {
		ee.Message = "invalid pin"
		return "", ee.ToError()
	}

	if r.ID == "" {
		ee.Message = "invalid agent id"
		return "", ee.ToError()
	}

	aId, err := primitive.ObjectIDFromHex(r.ID)

	if err != nil {
		ee.Error = err
		return "", ee.ToError()
	}

	ee.ObjectID = aId

	u := agent.Agent{ID: aId}
	err = u.Get(db)
	if err != nil {
		ee.Error = err
		ee.Message = "error getting agent"
		ee.Message += " - connecting ip: " + ip
		ee.Message += " - pin: " + r.PIN
		return "", ee.ToError()
	}

	if u.Pin != r.PIN {
		ee.Error = err
		ee.Message = "pins do not match"
		ee.Message += " - connecting ip: " + ip
		return "", ee.ToError()
	}

	session := Session{
		ID:      u.ID,
		IsAgent: true,
		IP:      ip,
	}

	err = session.Create(db)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to create session"
		ee.Message += " - connecting ip: " + ip
		return "", ee.ToError()
	}

	err = u.UpdateAgentVersion(r.AgentVersion, db)
	if err != nil {
		ee.Level = log.WarnLevel
		ee.Message = "unable to update agent version"
		ee.Print()
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
		ee.Message = "unable to generate signed string"
		ee.Error = err
		return "", ee.ToError()
	}

	out := map[string]any{
		"token": t,
		"data":  u,
	}

	bytes, err := json.Marshal(out)
	if err != nil {
		ee.Message = "unable to marshal output"
		ee.Error = err
		return "", ee.ToError()
	}

	return string(bytes), nil
}
