package web

import (
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"nw-guardian/internal/agent"
)

func addRouteProbes(r *Router) []*Route {
	var tempRoutes []*Route
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Probes",
		Path: "/probes/{agentid}",
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

			cId, err := primitive.ObjectIDFromHex(params.Get("check"))
			if err != nil {
				return ctx.JSON(err)
			}

			// todo handle edge cases? the user *could* break their install if not... hmmm...

			check := agent.Probe{ID: cId}

			// .Get will update it self instead of returning a list with a first object
			cc, err := check.Get(r.DB)
			if err != nil {
				return ctx.JSON(err)
			}

			log.Info(check)
			err = ctx.JSON(cc)
			if err != nil {
				return err
			}

			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "New Probe",
		Path: "/probes/new/{agentid}",
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

			aId, err := primitive.ObjectIDFromHex(params.Get("agentid"))
			if err != nil {
				return ctx.JSON(err)
			}

			// require check request
			req := agent.Probe{}
			err = ctx.ReadJSON(&req)
			if err != nil {
				return ctx.JSON(err)
			}
			req.Agent = aId

			// todo handle edge cases? the user *could* break their install if not... hmmm...

			err = req.Create(r.DB)
			if err != nil {
				return ctx.JSON(err)
			}

			ctx.StatusCode(http.StatusOK)

			return nil
		},
		Type: RouteType_POST,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Probe",
		Path: "/probes/{probeid}",
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

			cId, err := primitive.ObjectIDFromHex(params.Get("check"))
			if err != nil {
				return ctx.JSON(err)
			}

			// todo handle edge cases? the user *could* break their install if not... hmmm...

			check := agent.Probe{ID: cId}

			// .Get will update it self instead of returning a list with a first object
			_, err = check.Get(r.DB)
			if err != nil {
				return ctx.JSON(err)
			}

			log.Info(check)

			ctx.JSON(check)

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
