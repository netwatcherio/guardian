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

	AddAuthRoutes(r)

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
	for _, v := range r.Routes {
		// skip loading JWT for auth routes? will need to include the logout one otherwise it wouldn't do anything? or we just log out and ignore errors

		if !v.JWT && JWT {
			log.Warnf("Skipping %s - %s due to being NON-JWT route...", v.Name, v.Path)
			log.Warnf("Data - JWT: %v noJWT: %v", v.JWT, JWT)
			continue
		}

		if v.JWT && !JWT {
			log.Warnf("Skipping %s - %s due to being JWT route...", v.Name, v.Path)
			log.Warnf("Data - JWT: %v noJWT: %v", v.JWT, JWT)
			continue
		}

		log.Infof("Loaded route: %s (%s) - %s", v.Name, v.Type, v.Path)
		if v.Type == RouteType_GET {
			if err := r.App.Get(v.Path, func(c *fiber.Ctx) error {
				return c.SendString("Testing...")
			}); err != nil {
				// Handle the error here (e.g., log it)
				log.Errorf("Error setting up GET route: %v", err)
			}
		} else if v.Type == RouteType_POST {
			if err := r.App.Post(v.Path, func(c *fiber.Ctx) error {
				return v.Func(c)
			}); err != nil {
				// Handle the error here (e.g., log it)
				log.Errorf("Error setting up POST route: %v", err)
			}
		} else if v.Type == RouteType_WEBSOCKET {
			if err := r.App.Get("/ws", websocket.New(func(c *websocket.Conn) {
				v.FuncWS(c)
			})); err != nil {
				// Handle the error here (e.g., log it)
				log.Errorf("Error setting up WEBSOCKET route: %v", err)
			}
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
