package web

import "github.com/kataras/iris/v12"

func addRouteAgentAPI(r *Router) []*Route {
	var tempRoutes []*Route

	tempRoutes = append(tempRoutes, &Route{
		Name: "Agent API Login",
		Path: "/agent/login",
		JWT:  false,
		Func: func(ctx iris.Context) error {

			return nil
		},
		Type: RouteType_GET,
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
