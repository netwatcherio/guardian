package web

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"nw-guardian/internal/auth"
	"nw-guardian/internal/sites"
)

func addRouteSites(r *Router) []*Route {
	var tempRoutes []*Route
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Sites",
		Path: "/sites",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			t := ctx.Locals("user").(*jwt.Token)
			log.Warnf("%v", t)
			u, err := auth.GetUser(t, r.DB)
			if err != nil {
				return ctx.JSON(err)
			}

			getSites, err := sites.GetSites(u.ID, r.DB)
			if err != nil {
				return err
			}

			marshal, err := json.Marshal(getSites)
			if err != nil {
				return err
			}

			return ctx.Send(marshal)
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "New Site",
		Path: "/sites",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			ctx.Accepts("application/json") // "Application/json"
			t := ctx.Locals("user").(*jwt.Token)
			log.Warnf("%v", t)
			u, err := auth.GetUser(t, r.DB)
			if err != nil {
				return ctx.JSON(err)
			}

			s := new(sites.Site)
			if err := ctx.BodyParser(s); err != nil {
				return ctx.JSON(err)
			}

			err = s.Create(u.ID, r.DB)
			if err != nil {
				return ctx.JSON(err)
			}
			return ctx.SendStatus(fiber.StatusOK)
		},
		Type: RouteType_POST,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Site",
		Path: "/sites/:siteID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Delete Site",
		Path: "/sites/:siteID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "DELETE",
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Members",
		Path: "/sites/members",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Add Member",
		Path: "/sites/members",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: RouteType_POST,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Delete Member",
		Path: "/sites/members/:siteID/:userID",
		JWT:  true,
		Func: func(ctx *fiber.Ctx) error {
			return nil
		},
		Type: "DELETE",
	})

	return tempRoutes
}
