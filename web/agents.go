package web

import (
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"nw-guardian/internal/agent"
	"nw-guardian/internal/workspace"
)

func addRouteAgents(r *Router) []*Route {
	var tempRoutes []*Route

	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Agents for Workspace",
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

			aId, err := primitive.ObjectIDFromHex(params.Get("siteid"))
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			a := workspace.Workspace{ID: aId}
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
		Name: "Delete Agent",
		Path: "/agents/delete/{agentid}",
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
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			err = agent.DeleteAgent(r.DB, aId)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}
			ctx.StatusCode(http.StatusOK)
			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "New Agent for Workspace",
		Path: "/agents/new/{siteid}",
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

			s := workspace.Workspace{ID: sId}
			err = s.Get(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			cAgent := new(agent.Agent)
			ctx.ReadJSON(&cAgent)

			cAgent.Site = s.ID

			err = cAgent.Create(r.DB)
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
		Name: "Update Agent",
		Path: "/agents/update/{agentid}",
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

			sId, err := primitive.ObjectIDFromHex(params.Get("agentid"))
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			s := agent.Agent{ID: sId}

			cAgent := new(agent.Agent)
			err = ctx.ReadJSON(&cAgent)
			if err != nil {
				return err
			}

			cAgent.Site = s.ID

			err = cAgent.UpdateAgentDetails(r.DB, cAgent.Name, cAgent.Location, cAgent.PublicIPOverride)
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
		Name: "Deactivate an agent",
		Path: "/agents/deactivate/{agentid}",
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

			sId, err := primitive.ObjectIDFromHex(params.Get("agentid"))
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			s := agent.Agent{ID: sId}
			err = s.Get(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			err = s.Deactivate(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return err
			}

			ctx.StatusCode(http.StatusOK)

			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Agent",
		Path: "/agents/{agentid}",
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

			sId, err := primitive.ObjectIDFromHex(params.Get("agentid"))
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			s := agent.Agent{ID: sId}
			err = s.Get(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			ctx.JSON(s)

			return nil
		},
		Type: RouteType_GET,
	})

	return tempRoutes
}
