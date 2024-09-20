package web

import (
	"encoding/json"
	"errors"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"nw-guardian/internal/agent"
	"nw-guardian/internal/users"
	"nw-guardian/internal/workspace"
)

func addRouteSites(r *Router) []*Route {
	var tempRoutes []*Route

	tempRoutes = append(tempRoutes, &Route{
		Name: "Update Workspace",
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

			s := workspace.Workspace{ID: sId}

			cAgent := new(workspace.Workspace)
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

			getSites, err := workspace.GetSitesForMember(t.ID, r.DB)
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
		Name: "New Workspace",
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

			s := new(workspace.Workspace)
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
		Name: "Get MemberInfo",
		Path: "/sites/{siteid}/memberinfo",
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

			s := workspace.Workspace{ID: siteId}
			err = s.Get(r.DB)
			if err != nil {
				return ctx.JSON(err)
			}

			// Retrieve member info
			memberInfos, err := s.GetMemberInfos(r.DB)
			if err != nil {
				// Handle the error. Depending on your requirement, you might want to still return the site info without member details
				return ctx.JSON(err)
			}

			ctx.JSON(memberInfos)
			return nil
		},
		Type: RouteType_GET,
	})

	tempRoutes = append(tempRoutes, &Route{
		Name: "Workspace",
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

			s := workspace.Workspace{ID: siteId}
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
		Name: "Delete Workspace",
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
		Name: "Update Member Role",
		Path: "/sites/{siteid}/update_role",
		JWT:  true,
		Func: func(ctx iris.Context) error {

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

			info := workspace.MemberInfo{}

			err = ctx.ReadJSON(&info)
			if err != nil {
				return err
			}

			s := workspace.Workspace{ID: siteId}
			err = s.Get(r.DB)
			if err != nil {
				return err
			}

			if !s.IsMember(info.ID) {
				return errors.New("user is not a member of this site")
			}

			// Update the member's role
			err = s.UpdateMemberRole(info.ID, info.Role, r.DB)
			if err != nil {
				return err
			}

			ctx.StatusCode(http.StatusOK)

			return nil
		},
		Type: RouteType_POST, // POST is used for updating data
	})

	tempRoutes = append(tempRoutes, &Route{
		Name: "Add Member",
		Path: "/sites/{siteid}/invite",
		JWT:  true,
		Func: func(ctx iris.Context) error {

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

			info := workspace.MemberInfo{}

			err = ctx.ReadJSON(&info)
			if err != nil {
				return err
			}

			usr := users.User{Email: info.Email}
			uuu, err := usr.FromEmail(r.DB)
			if err != nil {
				// todo handle if no users exist with that email
				return err
			}

			info.ID = uuu.ID

			if info.Role == workspace.MemberRole_OWNER {
				return errors.New("only the owner can add owners")
			}

			s := workspace.Workspace{ID: siteId}
			err = s.Get(r.DB)
			if err != nil {
				return err
			}
			err = s.AddMember(info.ID, info.Role, r.DB)
			if err != nil {
				return err
			}

			return nil
		},
		Type: RouteType_POST,
	})
	tempRoutes = append(tempRoutes, &Route{
		Name: "Remove Member",
		Path: "/sites/{siteid}/remove",
		JWT:  true,
		Func: func(ctx iris.Context) error {

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

			info := workspace.MemberInfo{}

			err = ctx.ReadJSON(&info)
			if err != nil {
				return err
			}

			s := workspace.Workspace{ID: siteId}
			err = s.Get(r.DB)
			if err != nil {
				return err
			}

			if !s.IsMember(info.ID) {
				return errors.New("user is not a member of this site")
			}

			role, err := s.GetMemberRole(info.ID)
			if err != nil {
				return err
			}

			if role == workspace.MemberRole_OWNER {
				ctx.StatusCode(http.StatusInternalServerError)
				return errors.New("the owner cannot be removed")
			}

			err = s.RemoveMember(info.ID, r.DB)
			if err != nil {
				return err
			}

			ctx.StatusCode(http.StatusOK)

			return nil
		},
		Type: RouteType_POST, // or RouteType_DELETE if appropriate for your API design
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
		Name: "Workspace Groups",
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
