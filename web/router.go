package web

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
)

/**
/probes (GET) - List all probes
/probes (POST) - Create a new check
/probes/{checkID} (GET) - Get details for a specific check
/probes/{checkID} (DELETE) - Delete a specific check
/probes/{checkID}/data (GET) - Get data for a specific check

/agents (GET) - List all agents
/agents (POST) - Create a new agent
/agents/{agentID} (GET) - Get details for a specific agent
/agents/{agentID}/stats (GET) - Get general stats for a specific agent
/agents/{agentID} (DELETE) - Delete a specific agent
*/

type Router struct {
	App     *fiber.App
	Session *session.Store
	DB      *mongo.Database
	Routes  []Route
}

type Route struct {
	Name   string
	Path   string
	Type   string
	Func   RouteFunc
	FuncWS RouteFuncWS
}
type RouteFuncWS func(*websocket.Conn) error
type RouteFunc func(*fiber.Ctx) error

func NewRouter(mongoDB *mongo.Database) *Router {
	router := Router{
		App: fiber.New(),
		DB:  mongoDB,
	}
	return &router
}
func secretKey() jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		// Always check the signing method
		if t.Method.Alg() != jwtware.HS256 {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}

		signingKey := os.Getenv("KEY")

		return []byte(signingKey), nil
	}
}

var privateKey *rsa.PrivateKey

func (r *Router) Init() {

	if os.Getenv("DEBUG") != "" {
		log.Warning("Cross Origin requests allowed (ENV::DEBUG)")
		r.App.Use(cors.New())
	}

	var err error
	rng := rand.Reader
	privateKey, err = rsa.GenerateKey(rng, 2048)
	if err != nil {
		log.Fatalf("rsa.GenerateKey: %v", err)
	}

	// JWT Middleware
	r.App.Use(jwtware.New(jwtware.Config{
		KeyFunc: secretKey(),
	}))

	log.Info("Loading all routes...")
	if len(r.Routes) > 0 {
		log.Infof("Found %d route(s).", len(r.Routes))
		r.LoadRoutes()
	} else {
		log.Error("Failed to load routes. No routes found.")
	}
}

type RouteType string

const (
	RouteType_GET  RouteType = "GET"
	RouteType_POST RouteType = "POST"
)

func (r *Router) LoadRoutes() {
	for _, v := range r.Routes {
		log.Infof("Loaded route: %s (%s) - %s", v.Name, v.Type, v.Path)
		if v.Type == "GET" {
			r.App.Get(v.Path, func(c *fiber.Ctx) error {
				return v.Func(c)
			})
		} else if v.Type == "POST" {
			r.App.Post(v.Path, func(c *fiber.Ctx) error {
				return v.Func(c)
			})
		} else if v.Type == "WEBSOCKET" {
			// WebSocket route for authenticated users.
			r.App.Get("/ws", websocket.New(func(c *websocket.Conn) {
				v.FuncWS(c)
			}))
		}
	}
}

func (r *Router) Listen(host string) {
	err := r.App.Listen(host)
	if err != nil {
		return
	}
}
