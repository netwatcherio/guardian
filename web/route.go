package web

import (
	"github.com/kataras/iris/v12"
)

type Route struct {
	Name string
	Path string
	JWT  bool
	Type RouteType
	Func func(ctx iris.Context) error
}

type RouteType string

const (
	RouteType_GET       RouteType = "GET"
	RouteType_POST      RouteType = "POST"
	RouteType_WEBSOCKET RouteType = "WEBSOCKET"
)
