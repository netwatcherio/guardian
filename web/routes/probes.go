package routes

import (
	"github.com/gofiber/fiber/v2"
	"nw-guardian/web"
)

func AddProbesRoutes(r *web.Router) {
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
	r.Routes = append(r.Routes, web.Route{
		Name: "Get Probe",
		Path: "/probes/:probeID",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: web.RouteType_GET,
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "Delete Probe",
		Path: "/probes/:probeID",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "DELETE",
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "Get Probe Data",
		Path: "/probes/data/:probeID",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: web.RouteType_GET,
	})
}
