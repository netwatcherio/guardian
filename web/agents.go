package web

import (
	"github.com/gofiber/fiber/v2"
)

func addRouteAgents(r *Router) []*Route {
	var tempRoutes []*Route

	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Agents",
		Path: "/agents/:agentID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return ctx.SendString("Get Agents")
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "New Agents",
		Path: "/agents",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return ctx.SendString("Get Agents")
		},
		Type: RouteType_POST,
	})

	return tempRoutes
}
