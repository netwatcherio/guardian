package web

import (
	"github.com/kataras/iris/v12"
	"net/http"
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

			t, err := l.AgentLogin(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusUnauthorized)
				return nil
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
