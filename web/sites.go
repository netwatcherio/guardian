package web

import (
	"github.com/gofiber/fiber/v2"
)

func addRouteSites(r *Router) []*Route {
	var tempRoutes []*Route
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Sites",
		Path: "/sites",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "New Site",
		Path: "/sites",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_POST,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Site Agents",
		Path: "/sites/agents/:agentID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Site",
		Path: "/sites/:siteID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Delete Site",
		Path: "/sites/:siteID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "DELETE",
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Add Member",
		Path: "/sites/:siteID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Add Member",
		Path: "/sites/members",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_POST,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Delete Member",
		Path: "/sites/members/:siteID/:userID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "DELETE",
	})

	return tempRoutes
}
