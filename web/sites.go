package web

import (
	"github.com/gofiber/fiber/v2"
)

func AddSitesRoutes(r *Router) {
	r.Routes = append(r.Routes, &Route{
		Name: "Get Sites",
		Path: "/sites",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	r.Routes = append(r.Routes, &Route{
		Name: "New Site",
		Path: "/sites",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_POST,
	})
	r.Routes = append(r.Routes, &Route{
		Name: "Site Agents",
		Path: "/sites/agents/:agentID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	r.Routes = append(r.Routes, &Route{
		Name: "Site",
		Path: "/sites/:siteID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	r.Routes = append(r.Routes, &Route{
		Name: "Delete Site",
		Path: "/sites/:siteID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "DELETE",
	})
	r.Routes = append(r.Routes, &Route{
		Name: "Add Member",
		Path: "/sites/:siteID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	r.Routes = append(r.Routes, &Route{
		Name: "Add Member",
		Path: "/sites/members",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_POST,
	})
	r.Routes = append(r.Routes, &Route{
		Name: "Delete Member",
		Path: "/sites/members/:siteID/:userID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "DELETE",
	})
}
