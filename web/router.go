package web

import (
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

type Router struct {
	App     *fiber.App
	Session *session.Store
	DB      *mongo.Database
	Routes  []Route
}

type Route struct {
	Name   string
	Path   string
	Group  RouteGroup
	Type   RouteType
	Func   RouteFunc
	FuncWS RouteFuncWS
}

const (
	RouteGroup_AUTH RouteGroup = "AUTH"
)

type RouteGroup string
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

func (r *Router) Init() {

	if os.Getenv("DEBUG") != "" {
		log.Warning("Cross Origin requests allowed (ENV::DEBUG)")
		r.App.Use(cors.New())
	}

	log.Info("Loading all routes...")
	log.Infof("Found %d route(s).", len(r.Routes))
	if len(r.Routes) > 0 {
		r.LoadRoutes(true)
	} else {
		log.Error("Failed to no JWT routes. No routes found.")
	}

	r.App.Use(jwtware.New(jwtware.Config{
		KeyFunc: secretKey(),
	}))
	if len(r.Routes) > 0 {
		r.LoadRoutes(false)
	} else {
		log.Error("Failed to load routes. No routes found.")
	}

	// JWT Middleware TODO use cert or something more "static" in production
}

type RouteType string

const (
	RouteType_GET       RouteType = "GET"
	RouteType_POST      RouteType = "POST"
	RouteType_WEBSOCKET RouteType = "POST"
)

func (r *Router) LoadRoutes(noJwt bool) {
	for _, v := range r.Routes {
		// skip loading JWT for auth routes? will need to include the logout one otherwise it wouldn't do anything? or we just log out and ignore errors
		if noJwt {
			if v.Group != RouteGroup_AUTH {
				continue
			}
		} else {
			if v.Group == RouteGroup_AUTH {
				continue
			}
		}

		log.Infof("Loaded route: %s (%s) - %s", v.Name, v.Type, v.Path)
		if v.Type == RouteType_GET {
			r.App.Get(v.Path, func(c *fiber.Ctx) error {
				return v.Func(c)
			})
		} else if v.Type == RouteType_POST {
			r.App.Post(v.Path, func(c *fiber.Ctx) error {
				return v.Func(c)
			})
		} else if v.Type == RouteType_WEBSOCKET {
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
