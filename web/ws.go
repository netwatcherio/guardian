package web

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/websocket"
	"log"
	"net/http"
	"nw-guardian/internal/auth"
	"os"
	"strings"
	// Used when "enableJWT" constant is true:
)

func addWebSocketServer(r *Router) error {

	websocketServer := websocket.New(
		websocket.DefaultGorillaUpgrader, /* DefaultGobwasUpgrader can be used too. */
		getWebsocketEvents())

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

		return nil
	}

	r.WebSocketServer = websocketServer

	return nil
}

func getWebsocketEvents() websocket.Namespaces {
	serverEvents := websocket.Namespaces{
		"agent": websocket.Events{
			websocket.OnNamespaceConnected: func(nsConn *websocket.NSConn, msg websocket.Message) error {
				// with `websocket.GetContext` you can retrieve the Iris' `Context`.
				ctx := websocket.GetContext(nsConn.Conn)

				log.Printf("[%s] connected to namespace [%s] with IP [%s]",
					nsConn, msg.Namespace,
					ctx.RemoteAddr())
				return nil
			},
			websocket.OnNamespaceDisconnect: func(nsConn *websocket.NSConn, msg websocket.Message) error {
				// todo update agent status to be offline??
				log.Printf("[%s] disconnected from namespace [%s]", nsConn, msg.Namespace)
				return nil
			},
			"probe_get": func(nsConn *websocket.NSConn, msg websocket.Message) error {
				// room.String() returns -> NSConn.String() returns -> Conn.String() returns -> Conn.ID()
				log.Printf("[%s] sent: %s", nsConn, string(msg.Body))

				// Write message back to the client message owner with:
				// nsConn.Emit("chat", msg)
				// Write message to all except this client with:
				nsConn.Conn.Server().Broadcast(nsConn, msg)
				return nil
			},
		},
	}

	return serverEvents
}
