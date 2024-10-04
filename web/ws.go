package web

import (
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"nw-guardian/internal/agent"
	"nw-guardian/internal/auth"
	"os"
	"strings"
)

func addWebSocketServer(r *Router) error {

	websocketServer := websocket.New(
		websocket.DefaultGorillaUpgrader, /* DefaultGobwasUpgrader can be used too. */
		getWebsocketEvents(r))

	websocketServer.OnConnect = func(c *websocket.Conn) error {
		ctx := websocket.GetContext(c)

		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" {
			ctx.StatusCode(http.StatusUnauthorized)
			return errors.New("unauthorized. no token")
		}

		newToken := strings.ReplaceAll(tokenString, "Bearer ", "")

		token, err := jwt.Parse(newToken, func(token *jwt.Token) (interface{}, error) {
			// Here you should provide your JWT secret key
			return []byte(os.Getenv("KEY")), nil
		})

		if err != nil || !token.Valid {
			ctx.StatusCode(http.StatusUnauthorized)
			return errors.New("unauthorized. invalid token")
		}

		// todo change to get agent from token for auth & agent login to generate token??
		agent, err := auth.GetAgent(token, r.DB)
		if err != nil {
			ctx.StatusCode(http.StatusUnauthorized)
			return errors.New("unauthorized. invalid agent token")
		}

		log.Printf("This is an authenticated request\n")
		log.Printf("Agent: %v", agent.ID.String())

		log.Printf("[%s] connected to the server", c.ID())

		id, err := auth.GetSessionID(token)
		if err != nil {
			return err
		}

		ss := auth.Session{ID: agent.ID, SessionID: id, WSConn: c.ID()}
		err = ss.UpdateConnWS(r.DB)
		if err != nil {
			return err
		}

		return nil
	}

	r.WebSocketServer = websocketServer

	return nil
}

func getWebsocketEvents(r *Router) websocket.Namespaces {
	serverEvents := websocket.Namespaces{
		"agent": websocket.Events{
			websocket.OnNamespaceConnected: func(nsConn *websocket.NSConn, msg websocket.Message) error {
				// with `websocket.GetContext` you can retrieve the Iris' `Context`.
				//ctx := websocket.GetContext(nsConn.Conn)

				log.Infof("[%s] connected to namespace [%s]", /* with IP [%s]""*/
					nsConn, msg.Namespace, /*,
					ctx.Values().GetString("client_ip")*/)
				return nil
			},
			websocket.OnNamespaceDisconnect: func(nsConn *websocket.NSConn, msg websocket.Message) error {
				// todo update agent status to be offline??
				log.Infof("[%s] disconnected from namespace [%s]", nsConn, msg.Namespace)
				return nil
			},
			"probe_get": func(nsConn *websocket.NSConn, msg websocket.Message) error {
				// room.String() returns -> NSConn.String() returns -> Conn.String() returns -> Conn.ID()
				// log.Printf("[%s] sent: %s", nsConn, string(msg.Body))

				session, err := auth.GetSessionFromWSConn(nsConn.String(), r.DB)
				if err != nil {
					return err
				}

				a := agent.Agent{ID: session.ID}
				err = a.Get(r.DB)
				if err != nil {
					return err
				}

				err = a.UpdateTimestamp(r.DB)
				if err != nil {
					log.Error(err)
				}

				probe := agent.Probe{Agent: session.ID}
				// todo change this to build based on if the probe is an agent/group type probe
				// todo add group type probes ?? or just use agent type probes for groups??
				probes, err := probe.GetAllProbesForAgent(r.DB)
				if err != nil {
					log.Errorf(err.Error())
				}

				marshal, err := json.Marshal(probes)
				if err != nil {
					return err
				}

				nsConn.Emit("probe_get", marshal)

				// Write message back to the client message owner with:
				// nsConn.Emit("chat", msg)
				// Write message to all except this client with:
				//nsConn.Conn.Server().Broadcast(nsConn, msg)
				return nil
			},
			"probe_post": func(nsConn *websocket.NSConn, msg websocket.Message) error {
				// room.String() returns -> NSConn.String() returns -> Conn.String() returns -> Conn.ID()
				// log.Printf("[%s] sent: %s", nsConn, string(msg.Body))

				session, err := auth.GetSessionFromWSConn(nsConn.String(), r.DB)
				if err != nil {
					return err
				}

				a := agent.Agent{ID: session.ID}
				err = a.Get(r.DB)
				if err != nil {
					return err
				}

				data := agent.ProbeData{}

				err = json.Unmarshal(msg.Body, &data)
				if err != nil {
					log.Error(err)
					return err
				}

				r.ProbeDataChan <- data

				/*probe := agent.Probe{Agent: session.ID}
				probes, _ := probe.GetAll(r.DB)

				*/

				// Write message back to the client message owner with:
				// nsConn.Emit("chat", msg)
				// Write message to all except this client with:
				//nsConn.Conn.Server().Broadcast(nsConn, msg)
				return nil
			},
		},
	}

	return serverEvents
}
