package auth

import (
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
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
		return "", errors.New("invalid email address")
	}

	u := users.User{Email: r.Email}
	user, err := u.FromEmail(db)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(r.Password))
	if err != nil {
		return "", errors.New("invalid password, please ensure passwords match")
	}

	session := Session{
		ID:      user.ID,
		IsAgent: false,
	}

	err = session.Create(db)
	if err != nil {
		return "", err
	}

	// Create the Claims
	claims := jwt.MapClaims{
		"item_id":    session.ID.Hex(),
		"session_id": session.ID.Hex(),
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
		"item":  *user,
	}

	bytes, err := json.Marshal(out)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

type Register struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
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
		// todo validate email
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
		"session_id": session.ID.Hex(),
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
		"item":  *out,
	}

	bytes, err := json.Marshal(outMap)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
