package web

func addRouteAgentWS(r *Router) []*Route {
	var tempRoutes []*Route

	tempRoutes = append(tempRoutes, &Route{
		Name: "Agent WebSocket",
		Path: "/agent/ws",
		JWT:  true,
		/*FuncWS: func(conn *websocket.Conn) error {
			// Read the JWT token from the request header.

			t := conn.Locals("item").(*jwt.Token)
			_, err := auth.GetUser(t, r.DB)
			if err != nil {
				return conn.WriteJSON(err)
			}

			if err != nil {
				log.Error(err)
			}

			// You can access claims like user ID from the token.
			claims := t.Claims.(jwt.MapClaims)
			userID := int(claims["item_id"].(float64))

			// Simulate user-specific WebSocket communication.
			log.Printf("User %d connected via WebSocket", userID)

			// Handle WebSocket communication.
			for {
				// Read a message from the client.
				messageType, p, err := conn.ReadMessage()
				if err != nil {
					break
				}

				// Handle different message types (e.g., text, binary).
				// You can perform actions based on the message type.
				switch messageType {
				case websocket.TextMessage:
					log.Printf("Received message from User %d: %s", userID, string(p))
					// Handle text message (p is a []byte containing the message).
					// You can parse the message and send a response.
					// Example:
					// handleTextMessage(c, p)
				case websocket.BinaryMessage:
					// Handle binary message (p is a []byte containing the message).
					// Example:
					// handleBinaryMessage(c, p)
				}
			}

			log.Printf("User %d disconnected from WebSocket", userID)
			return nil
		},*/
		Type: RouteType_WEBSOCKET,
	})

	return tempRoutes
}
