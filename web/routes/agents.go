package routes

import (
	"github.com/gofiber/fiber/v2"
	"nw-guardian/web"
)

func AddAgentsRoutes(r *web.Router) {
	r.Routes = append(r.Routes, web.Route{
		Name: "Get Probes",
		Path: "/probes/:agentID",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: web.RouteType_GET,
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "New Probe",
		Path: "/probes",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: web.RouteType_POST,
	})
}
