package web

import (
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"nw-guardian/internal/agent"
	"nw-guardian/internal/auth"
)

func addRouteAgentAPI(r *Router) []*Route {
	var tempRoutes []*Route

	tempRoutes = append(tempRoutes, &Route{
		Name: "Agent API Login",
		Path: "/agent/login",
		JWT:  false,
		Func: func(ctx iris.Context) error {

			ctx.ContentType("application/json") // "Application/json"

			var l auth.AgentLogin
			err := ctx.ReadJSON(&l)
			if err != nil {
				ctx.StatusCode(http.StatusBadRequest)
				return nil
			}

			t, err := l.AgentLogin(ctx.Values().GetString("client_ip"), r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusUnauthorized)
				return nil
			}

			hex, err := primitive.ObjectIDFromHex(l.ID)
			if err != nil {
				return err
			}

			a := agent.Agent{ID: hex}
			err = a.UpdateTimestamp(r.DB)
			if err != nil {
				log.Error(err)
			}
			err = a.Initialize(r.DB)
			if err != nil {
				return err
			}

			_, err = ctx.Write([]byte(t))
			if err != nil {
				return err
			}
			return nil
		},
		Type: RouteType_POST,
	})

	tempRoutes = append(tempRoutes, &Route{
		Name: "Agent API - Get Probes",
		Path: "/agent/probes",
		JWT:  true,
		Func: func(ctx iris.Context) error {

			return nil
		},
		Type: RouteType_GET,
	})

	return tempRoutes
}
