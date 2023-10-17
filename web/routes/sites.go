package routes

import (
	"github.com/gofiber/fiber/v2"
	"nw-guardian/web"
)

/*
/sites (GET) - List all sites TODO
/sites/agents/{siteID} - (Get) Gets all agents for specific site TODO
/sites (POST) - Create a new site TODO
/sites/{siteID} (GET) - Get details for a specific site TODO
/sites/{siteID} (DELETE) - Delete a specific site TODO
/sites/members/{siteID} (POST) - Add a member to a specific site TODO
/sites/members/{siteID}/{memberID} (DELETE) - Delete a member to a specific site TODO
*/

func AddSitesRoutes(r *web.Router) {
	r.Routes = append(r.Routes, web.Route{
		Name: "Get Sites",
		Path: "/sites",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "GET",
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "New Site",
		Path: "/sites",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "POST",
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "Site Agents",
		Path: "/sites/agents/:agentID",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "GET",
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "Site",
		Path: "/sites/:siteID",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "GET",
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
		Type: "GET",
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "Add Member",
		Path: "/sites/members",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "POST",
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
