package web

import (
	"encoding/json"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"nw-guardian/internal/agent"
	"nw-guardian/internal/site"
)

func addRouteSites(r *Router) []*Route {
	var tempRoutes []*Route

	tempRoutes = append(tempRoutes, &Route{
		Name: "Update Site",
		Path: "/sites/update/{siteid}",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			ctx.ContentType("application/json") // "Application/json"
			t := GetClaims(ctx)
			_, err := t.FromID(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			params := ctx.Params()

			sId, err := primitive.ObjectIDFromHex(params.Get("siteid"))
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			s := site.Site{ID: sId}

			cAgent := new(site.Site)
			ctx.ReadJSON(&cAgent)

			err = s.UpdateSiteDetails(r.DB, cAgent.Name, cAgent.Location, cAgent.Description)
			if err != nil {
				log.Error(err)
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			ctx.StatusCode(http.StatusOK)

			return nil
		},
		Type: RouteType_POST,
	})

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

			getSites, err := site.GetSitesForMember(t.ID, r.DB)
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

			s := new(site.Site)
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
			ctx.ContentType("application/json") // "Application/json"
			t := GetClaims(ctx)
			_, err := t.FromID(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			params := ctx.Params()
			siteId, err := primitive.ObjectIDFromHex(params.Get("siteid"))
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			s := site.Site{ID: siteId}
			err = s.Get(r.DB)
			if err != nil {
				return ctx.JSON(err)
			}
			ctx.JSON(s)
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
	tempRoutes = append(tempRoutes, &Route{
		Name: "New Agent Group",
		Path: "/sites/{siteid}/groups",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			ctx.ContentType("application/json") // "Application/json"
			t := GetClaims(ctx)
			_, err := t.FromID(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			params := ctx.Params()
			siteId, err := primitive.ObjectIDFromHex(params.Get("siteid"))
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			s := new(agent.Group)
			s.SiteID = siteId
			err = ctx.ReadJSON(&s)
			if err != nil {
				return err
			}

			err = s.Create(r.DB)
			if err != nil {
				return ctx.JSON(err)
			}
			ctx.StatusCode(http.StatusOK)
			return nil
		},
		Type: RouteType_POST,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Site Groups",
		Path: "/sites/{siteid}/groups",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			ctx.ContentType("application/json") // "Application/json"
			t := GetClaims(ctx)
			_, err := t.FromID(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			params := ctx.Params()
			siteId, err := primitive.ObjectIDFromHex(params.Get("siteid"))
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			s := agent.Group{SiteID: siteId}
			groups, err := s.GetAll(r.DB)
			if err != nil {
				return ctx.JSON(err)
			}
			err = ctx.JSON(groups)
			if err != nil {
				return err
			}
			return nil
		},
		Type: RouteType_GET,
	})

	return tempRoutes
}
