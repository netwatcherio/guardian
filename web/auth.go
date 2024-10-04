package web

import (
	"github.com/kataras/iris/v12"
	"net/http"
	"nw-guardian/internal/auth"
)

func addRouteAuth(r *Router) []*Route {
	var tempRoutes []*Route

	tempRoutes = append(tempRoutes, &Route{
		Name: "Login",
		Path: "/auth/login",
		JWT:  false,
		Func: func(ctx iris.Context) error {
			ctx.ContentType("application/json") // "Application/json"

			var l auth.Login
			err := ctx.ReadJSON(&l)
			if err != nil {
				ctx.StatusCode(http.StatusBadRequest)
				return nil
			}

			t, err := l.Login(ctx.Values().GetString("client_ip"), r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusUnauthorized)
				return nil
			}
			_, err = ctx.Write([]byte(t))
			if err != nil {
				return err
			}
			return nil

		},
		Type: RouteType_POST,
	})

	tempRoutes = append(tempRoutes, &Route{
		Name: "Register",
		Path: "/auth/register",
		JWT:  false,
		Func: func(ctx iris.Context) error {
			ctx.ContentType("Application/json") // "Application/json"

			var reg auth.Register
			err := ctx.ReadJSON(&reg)
			if err != nil {
				ctx.StatusCode(http.StatusBadRequest)
				return err
			}
			t, err := reg.Register(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusConflict)
				return err
			}
			_, err = ctx.Write([]byte(t))
			if err != nil {
				return err
			}

			return nil
		},
		Type: RouteType_POST,
	})

	return tempRoutes
}
