package router

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"nw-guardian/internal/auth"
)

func (r *Router) login() {
	r.App.Post("/auth/login", func(c *fiber.Ctx) error {
		c.Accepts("application/json") // "Application/json"

		var l auth.Login
		err := json.Unmarshal(c.Body(), &l)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return nil
		}

		t, err := l.Login(r.DB)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			return nil
		}

		return c.Send([]byte(t))
	})
}

func (r *Router) register() {
	r.App.Post("/auth/register", func(c *fiber.Ctx) error {
		c.Accepts("Application/json") // "Application/json"

		var reg auth.Register
		err := json.Unmarshal(c.Body(), &reg)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return err
		}
		t, err := reg.Register(r.DB)
		if err != nil {
			c.Status(http.StatusConflict)
			return err
		}

		return c.Send([]byte(t))
	})
}
