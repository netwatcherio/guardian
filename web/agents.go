package web

import (
	"github.com/gofiber/fiber/v2"
)

func addRouteAgents(r *Router) {
	r.Routes = append(r.Routes, &Route{
		Name: "Get Agents",
		Path: "/agents/:agentID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	r.Routes = append(r.Routes, &Route{
		Name: "New Agents",
		Path: "/agents",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_POST,
	})
}
