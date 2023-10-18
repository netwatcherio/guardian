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
	Routes  []*Route
}

func NewRouter(mongoDB *mongo.Database) *Router {
	router := &Router{
		App: fiber.New(),
		DB:  mongoDB,
	}
	return router
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

	r.Routes = append(r.Routes, addRouteAuth(r)...)
	r.Routes = append(r.Routes, addRouteAgents(r)...)
	r.Routes = append(r.Routes, addRouteSites(r)...)
	r.Routes = append(r.Routes, addRouteAgentWS(r)...)
	r.Routes = append(r.Routes, addRouteProbes(r)...)

	if os.Getenv("DEBUG") != "" {
		log.Warning("Cross Origin requests allowed (ENV::DEBUG)")
		r.App.Use(cors.New())
	}

	log.Info("Loading all routes...")
	log.Infof("Found %d route(s).", len(r.Routes))
	if len(r.Routes) > 0 {
		log.Info("Skipping routes that require JWT...")
		r.LoadRoutes(false)

		log.Info("Enabling JWT Middleware...")
		r.App.Use(jwtware.New(jwtware.Config{
			KeyFunc: secretKey(),
		}))

		log.Info("Loading JWT routes...")
		r.LoadRoutes(true)
	} else {
		log.Error("Failed to no JWT routes. No routes found.")
	}

	// JWT Middleware TODO use cert or something more "static" in production
}

func (r *Router) LoadRoutes(JWT bool) {
	for n, _ := range r.Routes {
		// skip loading JWT for auth routes? will need to include the logout one otherwise it wouldn't do anything? or we just log out and ignore errors

		v := r.Routes[n]

		if !v.JWT && JWT {
			log.Warnf("JWT route... SKIP... %s - %s", v.Name, v.Path)
			continue
		}

		if v.JWT && !JWT {
			log.Warnf("not JWT route... SKIP... %s - %s", v.Name, v.Path)
			continue
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
			r.App.Get("/ws", websocket.New(func(c *websocket.Conn) {
				v.FuncWS(c)
			}))
		}

	}
}

func (r *Router) Listen(host string) {
	err := r.App.Listen(host)
	if err != nil {
		log.Error(err)
		return
	}
}
