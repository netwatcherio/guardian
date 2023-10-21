package web

import "github.com/kataras/iris/v12"

func addRouteProbes(r *Router) []*Route {
	var tempRoutes []*Route
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Probes",
		Path: "/probes/{agentid}",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "New Probe",
		Path: "/probes",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			return nil
		},
		Type: RouteType_POST,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Probe",
		Path: "/probes/{probeid}",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Delete Probe",
		Path: "/probes/{probeid}",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			return nil
		},
		Type: "DELETE",
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Probe Data",
		Path: "/probes/data/{probeid}",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			return nil
		},
		Type: RouteType_GET,
	})
	return tempRoutes
}
