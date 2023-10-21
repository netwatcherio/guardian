package web

import (
	"github.com/kataras/iris/v12"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"nw-guardian/internal/sites"
)

func addRouteAgents(r *Router) []*Route {
	var tempRoutes []*Route

	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Agents for Site",
		Path: "/agents/site/{siteid}",
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

			aId, err := primitive.ObjectIDFromHex(params.Get("siteID"))
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			a := sites.Site{ID: aId}
			err = a.Get(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			agents, err := a.GetAgents(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			return ctx.JSON(agents)
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "New Agent for Site",
		Path: "/agents/new/{siteid}",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			return nil
		},
		Type: RouteType_POST,
	})

	return tempRoutes
}
