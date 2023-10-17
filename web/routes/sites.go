package routes

import (
	"github.com/gofiber/fiber/v2"
	"nw-guardian/web"
)

func AddSitesRoutes(r *web.Router) {
	r.Routes = append(r.Routes, web.Route{
		Name: "Get Sites",
		Path: "/sites",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: web.RouteType_GET,
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "New Site",
		Path: "/sites",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: web.RouteType_POST,
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "Site Agents",
		Path: "/sites/agents/:agentID",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: web.RouteType_GET,
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "Site",
		Path: "/sites/:siteID",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: web.RouteType_GET,
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "Delete Site",
		Path: "/sites/:siteID",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "DELETE",
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "Add Member",
		Path: "/sites/:siteID",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: web.RouteType_GET,
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "Add Member",
		Path: "/sites/members",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: web.RouteType_POST,
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "Delete Member",
		Path: "/sites/members/:siteID/:userID",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "DELETE",
	})
}
