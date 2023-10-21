package web

import (
	"encoding/json"
	"github.com/kataras/iris/v12"
	"net/http"
	"nw-guardian/internal/sites"
)

func addRouteSites(r *Router) []*Route {
	var tempRoutes []*Route
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Sites",
		Path: "/sites",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			t := GetClaims(ctx)
			_, err := t.FromID(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			getSites, err := sites.GetSitesForMember(t.ID, r.DB)
			if err != nil {
				return err
			}

			marshal, err := json.Marshal(getSites)
			if err != nil {
				return err
			}

			ctx.Write(marshal)

			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "New Site",
		Path: "/sites",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			ctx.ContentType("application/json") // "Application/json"
			t := GetClaims(ctx)
			_, err := t.FromID(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			s := new(sites.Site)
			err = ctx.ReadJSON(&s)
			if err != nil {
				return err
			}

			err = s.Create(t.ID, r.DB)
			if err != nil {
				return ctx.JSON(err)
			}
			ctx.StatusCode(http.StatusOK)
			return nil
		},
		Type: RouteType_POST,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Site",
		Path: "/sites/{siteid}",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Delete Site",
		Path: "/sites/{siteid}",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			return nil
		},
		Type: "DELETE",
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Members",
		Path: "/sites/members",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Add Member",
		Path: "/sites/members",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			return nil
		},
		Type: RouteType_POST,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Delete Member",
		Path: "/sites/members/{siteid}/{userid}",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			return nil
		},
		Type: "DELETE",
	})

	return tempRoutes
}
