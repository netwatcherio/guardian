package routes

import (
	"github.com/gofiber/fiber/v2"
	"nw-guardian/web"
)

/*
/agents (GET) - List all agents TODO
/agents (POST) - Create a new agent TODO
/agents/{agentID} (GET) - Get details for a specific agent TODO
/agents/{agentID}/stats (GET) - Get general stats for a specific agent TODO
/agents/{agentID} (DELETE) - Delete a specific agent TODO
*/

func AddAgentsRoutes(r *web.Router) {
	r.Routes = append(r.Routes, web.Route{
		Name: "Get Probes",
		Path: "/probes/:agentID",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "GET",
	})
	r.Routes = append(r.Routes, web.Route{
		Name: "New Probe",
		Path: "/probes",
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "POST",
	})
}
