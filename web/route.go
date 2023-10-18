package web

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type Route struct {
	Name   string
	Path   string
	JWT    bool
	Type   RouteType
	Func   func(*fiber.Ctx) error
	FuncWS func(*websocket.Conn) error
}

type RouteType string

const (
	RouteType_GET       RouteType = "GET"
	RouteType_POST      RouteType = "POST"
	RouteType_WEBSOCKET RouteType = "WEBSOCKET"
)

func initRoutes(r *Router) {
	addRouteAuth(r)
	addRouteSites(r)
	addRouteAgents(r)
	addRouteProbes(r)
	addRouteAgentWS(r)
}
