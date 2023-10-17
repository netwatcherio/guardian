package routes

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"nw-guardian/internal/auth"
	"nw-guardian/web"
)

/*
/auth/register (POST) - User registration
/auth/login (POST) - User login
/auth/logout (POST) - User logout
/auth/password-reset (POST) - Request a password reset
/auth/password-reset/{token} (POST) - Reset the password using a reset token
/auth/token-refresh (POST) - Refresh a JWT token (if using JWT-based authentication)
/auth/profile (GET) - Get the user's profile
/auth/profile (PUT) - Update the user's profile
/auth/profile/picture (POST) - Upload a profile picture
*/

func AddAuthRoutes(r *web.Router) {
	r.Routes = append(r.Routes, web.Route{
		Name: "Login",
		Path: "/auth/login",
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
		Path: "/auth/register",
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
