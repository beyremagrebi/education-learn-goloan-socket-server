package web

import (
	"fmt"

	socketio "github.com/googollee/go-socket.io"
)

// RegisterEvents registers web-related events
func RegisterEvents(server *socketio.Server) {
	server.OnEvent("/", "send-meet-notification", func(s socketio.Conn, users any) {
		fmt.Printf("Hello meet notification: %v\n", users)
		s.Emit("send-meet-notification", users)
		server.BroadcastToNamespace("/", "send-meet-notification", users)
	})
}
