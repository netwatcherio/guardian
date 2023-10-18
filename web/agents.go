package web

import (
	"github.com/gofiber/fiber/v2"
)

func addRouteAgents(r *Router) []*Route {
	var tempRoutes []*Route

	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Agents for Site",
		Path: "/agents/site/:siteID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return ctx.SendString("Get Agents")
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "New Agent for Site",
		Path: "/agents/new/:siteID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return ctx.SendString("Get Agents")
		},
		Type: RouteType_POST,
	})

	return tempRoutes
}
