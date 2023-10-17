package routes

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"nw-guardian/internal/auth"
	"nw-guardian/web"
)

func AddAuthRoutes(r *web.Router) {
	r.Routes = append(r.Routes, web.Route{
		Name: "Login",
		Path: "/login",
		Func: func(ctx *fiber.Ctx) error {
			ctx.Accepts("application/json") // "Application/json"

			var l auth.Login
			err := json.Unmarshal(ctx.Body(), &l)
			if err != nil {
				ctx.Status(http.StatusBadRequest)
				return nil
			}

			t, err := l.Login(r.DB)
			if err != nil {
				ctx.Status(http.StatusUnauthorized)
				return nil
			}

			return ctx.Send([]byte(t))
		},
		Type: "POST",
	})

	r.Routes = append(r.Routes, web.Route{
		Name: "Register",
		Path: "/register",
		Func: func(ctx *fiber.Ctx) error {
			ctx.Accepts("Application/json") // "Application/json"

			var reg auth.Register
			err := json.Unmarshal(ctx.Body(), &reg)
			if err != nil {
				ctx.Status(http.StatusBadRequest)
				return err
			}
			t, err := reg.Register(r.DB)
			if err != nil {
				ctx.Status(http.StatusConflict)
				return err
			}

			return ctx.Send([]byte(t))
		},
		Type: "POST",
	})
}
