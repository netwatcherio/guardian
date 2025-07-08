package web

import (
	"bytes"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io/ioutil"
	"net/http"
	"nw-guardian/internal/agent"
)

func addRouteProbes(r *Router) []*Route {
	var tempRoutes []*Route
	tempRoutes = append(tempRoutes, &Route{
		Name: "Delete Probe",
		Path: "/probe/delete/{pid}",
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

			aId, err := primitive.ObjectIDFromHex(params.Get("pid"))
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			p := agent.Probe{ID: aId}
			err = p.Delete(r.DB)
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
		Name: "Get NetworkInfo Probe",
		Path: "/netinfo/{agentid}",
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

			cId, err := primitive.ObjectIDFromHex(params.Get("agentid"))
			if err != nil {
				return ctx.JSON(err)
			}

			// todo handle edge cases? the user *could* break their install if not... hmmm...

			check := agent.Probe{Agent: cId, Type: agent.ProbeType_NETWORKINFO}

			// .Get will update it self instead of returning a list with a first object
			dd, err := check.Get(r.DB)
			if err != nil {
				return ctx.JSON(err)
			}

			dd[0].Agent = primitive.ObjectID{0}

			data, err := dd[0].GetData(&agent.ProbeDataRequest{Recent: true, Limit: 1}, r.DB)
			if err != nil {
				return err
			}

			// todo only return first element, we don't care currently about previous IPs and such...
			err = ctx.JSON(data[len(data)-1])
			if err != nil {
				return err
			}

			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get System Info Probe",
		Path: "/sysinfo/{agentid}",
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

			cId, err := primitive.ObjectIDFromHex(params.Get("agentid"))
			if err != nil {
				return ctx.JSON(err)
			}

			// todo handle edge cases? the user *could* break their install if not... hmmm...

			check := agent.Probe{Agent: cId, Type: agent.ProbeType_SYSTEMINFO}

			// .Get will update it self instead of returning a list with a first object
			dd, err := check.Get(r.DB)
			if err != nil {
				return ctx.JSON(err)
			}

			dd[0].Agent = primitive.ObjectID{0}

			data, err := dd[0].GetData(&agent.ProbeDataRequest{Recent: true, Limit: 1}, r.DB)
			if err != nil {
				return err
			}

			// todo only return first element, we don't care currently about previous IPs and such...
			err = ctx.JSON(data[len(data)-1])
			if err != nil {
				return err
			}

			return nil
		},
		Type: RouteType_GET,
	})

	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Similar Probes",
		Path: "/probes/similar/{probeid}",
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

			cId, err := primitive.ObjectIDFromHex(params.Get("probeid"))
			if err != nil {
				return ctx.JSON(err)
			}

			// todo handle edge cases? the user *could* break their install if not... hmmm...

			check := agent.Probe{ID: cId}

			// this actually returns the probes isntead of updating now
			cc, err := check.Get(r.DB)
			if err != nil {
				return ctx.JSON(err)
			}

			probes, err := cc[0].FindSimilarProbes(r.DB)
			if err != nil {
				return err
			}

			err = ctx.JSON(probes)
			if err != nil {
				return err
			}

			return nil
		},
		Type: RouteType_GET,
	})
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

			cId, err := primitive.ObjectIDFromHex(params.Get("agentid"))
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

			//log.Info(check)
			err = ctx.JSON(&cc)
			if err != nil {
				return err
			}

			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get Probe",
		Path: "/probe/{probeId}",
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

			cId, err := primitive.ObjectIDFromHex(params.Get("probeId"))
			if err != nil {
				return ctx.JSON(err)
			}

			// todo handle edge cases? the user *could* break their install if not... hmmm...

			check := agent.Probe{ID: cId}

			// .Get will update it self instead of returning a list with a first object
			cc, err := check.Get(r.DB)
			if err != nil {
				return err
			}

			//log.Info(check)
			err = ctx.JSON(&cc)
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
			cc, err := check.Get(r.DB)
			if err != nil {
				return ctx.JSON(err)
			}

			err = ctx.JSON(cc)
			if err != nil {
				return err
			}

			return nil
		},
		Type: RouteType_GET,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Get All Probes",
		Path: "/probes/agent/{agent}",
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

			cId, err := primitive.ObjectIDFromHex(params.Get("agent"))
			if err != nil {
				return ctx.JSON(err)
			}

			// todo handle edge cases? the user *could* break their install if not... hmmm...

			check := agent.Probe{Agent: cId}

			// .Get will update it self instead of returning a list with a first object
			probes, err := check.GetAll(r.DB)
			if err != nil {
				return ctx.JSON(err)
			}

			//log.Info(check)

			err = ctx.JSON(probes)
			if err != nil {
				return err
			}

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
		Name: "Get Probe Agent Data",
		Path: "/probes/data/{probe}",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			ctx.ContentType("application/json")

			t := GetClaims(ctx)
			_, err := t.FromID(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			params := ctx.Params()

			cId, err := primitive.ObjectIDFromHex(params.Get("probe"))
			if err != nil {
				return ctx.JSON(err)
			}

			probe := agent.Probe{ID: cId}

			pp, err := probe.Get(r.DB)
			if err != nil || len(pp) == 0 {
				return ctx.JSON(map[string]string{"error": "probe not found"})
			}

			pp[0].Agent = primitive.ObjectID{0}

			// Read raw body for logging (keeping your debug code)
			rawBody, _ := ioutil.ReadAll(ctx.Request().Body)
			log.Info("Raw request body: ", string(rawBody))

			// Recreate the body for ReadJSON
			ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(rawBody))

			req := agent.ProbeDataRequest{}
			err = ctx.ReadJSON(&req)
			if err != nil {
				log.Errorf("Error reading JSON: %s", err)
				return err
			}

			// NEW: Check if this is an AGENT probe or if grouping is requested
			if pp[0].Type == "AGENT" || ctx.URLParam("grouped") == "true" {
				// Check format parameter
				format := ctx.URLParam("format")

				switch format {
				case "flat":
					// Return flat array format
					data, err := pp[0].GetAgentProbeDataFlat(&req, r.DB)
					if err != nil {
						return err
					}
					return ctx.JSON(data)

				case "simple":
					// Return simple map format (type -> data)
					data, err := pp[0].GetAgentProbeData(&req, r.DB)
					if err != nil {
						return err
					}
					return ctx.JSON(data)

				default:
					// Return full grouped format (default for AGENT probes)
					data, err := pp[0].GetAgentProbeDataGrouped(&req, r.DB)
					if err != nil {
						return err
					}
					return ctx.JSON(data)
				}
			}

			// Original behavior for non-AGENT probes
			get, err := pp[0].GetData(&req, r.DB)
			if err != nil {
				return err
			}

			err = ctx.JSON(get)
			if err != nil {
				return err
			}

			return nil
		},

		Type: RouteType_POST,
	})
	/*tempRoutes = append(tempRoutes, &Route{
		Name: "Get Probe Data",
		Path: "/probes/data/{probe}",
		JWT:  true,
		Func: func(ctx iris.Context) error {
			ctx.ContentType("application/json") // "Application/json"

			// Print content type for debugging
			//log.Info("Content Type: ", ctx.GetHeader("Content-Type"))

			t := GetClaims(ctx)
			_, err := t.FromID(r.DB)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			params := ctx.Params()

			cId, err := primitive.ObjectIDFromHex(params.Get("probe"))
			if err != nil {
				return ctx.JSON(err)
			}

			probe := agent.Probe{ID: cId}

			pp, err := probe.Get(r.DB)
			pp[0].Agent = primitive.ObjectID{0}

			// Read raw body for logging
			rawBody, _ := ioutil.ReadAll(ctx.Request().Body)
			log.Info("Raw request body: ", string(rawBody))

			// Recreate the body for ReadJSON
			ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(rawBody))

			req := agent.ProbeDataRequest{}
			err = ctx.ReadJSON(&req)
			if err != nil {
				log.Errorf("Error reading JSON: %s", err)
				return err
			}

			get, err := pp[0].GetData(&req, r.DB)
			if err != nil {
				return err
			}

			err = ctx.JSON(get)
			if err != nil {
				return err
			}

			return nil
		},

		Type: RouteType_POST,
	})*/
	tempRoutes = append(tempRoutes, &Route{
		Name: "Update First Probe Target",
		Path: "/probe/first_target_update/{probeid}", // fuck i think im braindead
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

			sId, err := primitive.ObjectIDFromHex(params.Get("probeid"))
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return nil
			}

			s := agent.Probe{ID: sId}

			req := agent.Probe{}
			err = ctx.ReadJSON(&req)
			if err != nil {
				return ctx.JSON(err)
			}

			err = s.UpdateFirstProbeTarget(r.DB, req.Config.Target[0].Target)
			if err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				return err
			}

			ctx.StatusCode(http.StatusOK)

			return nil
		},
		Type: RouteType_POST,
	})
	return tempRoutes
}
