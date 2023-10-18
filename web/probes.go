package web

import (
	"github.com/gofiber/fiber/v2"
)

func addRouteProbes(r *Router) {
	r.Routes = append(r.Routes, &Route{
		Name: "Get Probes",
		Path: "/probes/:agentID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	r.Routes = append(r.Routes, &Route{
		Name: "New Probe",
		Path: "/probes",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_POST,
	})
	r.Routes = append(r.Routes, &Route{
		Name: "Get Probe",
		Path: "/probes/:probeID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	r.Routes = append(r.Routes, &Route{
		Name: "Delete Probe",
		Path: "/probes/:probeID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "DELETE",
	})
	r.Routes = append(r.Routes, &Route{
		Name: "Get Probe Data",
		Path: "/probes/data/:probeID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
}
