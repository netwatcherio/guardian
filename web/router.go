package web

import (
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type Router struct {
	App    *iris.Application
	DB     *mongo.Database
	Routes []*Route
}

func NewRouter(mongoDB *mongo.Database) *Router {
	router := &Router{
		App: iris.New(),
		DB:  mongoDB,
	}
	return router
}

func (r *Router) Init() {

	r.Routes = append(r.Routes, addRouteAuth(r)...)
	r.Routes = append(r.Routes, addRouteAgents(r)...)
	r.Routes = append(r.Routes, addRouteSites(r)...)
	r.Routes = append(r.Routes, addRouteAgentAPI(r)...)
	r.Routes = append(r.Routes, addRouteProbes(r)...)

	log.Info("Loading all routes...")
	log.Infof("Found %d route(s).", len(r.Routes))
	if len(r.Routes) > 0 {
		log.Info("Skipping routes that require JWT...")
		r.LoadRoutes(false)

		log.Info("Enabling JWT Middleware...")
		// TODO JWT middlewear
		r.App.Use(VerifySession())

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
			r.App.Get(v.Path, func(ctx iris.Context) {
				v.Func(ctx)
			})
		} else if v.Type == RouteType_POST {
			r.App.Post(v.Path, func(ctx iris.Context) {
				v.Func(ctx)
			})
		} else if v.Type == RouteType_WEBSOCKET {
			/*r.App.Get("/ws", websocket.New(func(c *websocket.Conn) {
				log.Info("SWITCH TO HTTP/2")
			}))*/
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
