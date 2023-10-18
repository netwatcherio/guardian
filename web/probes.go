package web

import (
	"github.com/gofiber/fiber/v2"
)

func addRouteProbes(r *Router) []*Route {
	var tempRoutes []*Route
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Probes",
		Path: "/probes/:agentID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "New Probe",
		Path: "/probes",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_POST,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Probe",
		Path: "/probes/:probeID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Delete Probe",
		Path: "/probes/:probeID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "DELETE",
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Probe Data",
		Path: "/probes/data/:probeID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	return tempRoutes
}
