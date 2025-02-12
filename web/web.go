package web

import (
	socketio "github.com/googollee/go-socket.io"
)

// RegisterEvents registers web-related events
func RegisterEvents(server *socketio.Server) {
	server.OnEvent("/", "send-meet-notification", func(s socketio.Conn, users any) {
		server.BroadcastToNamespace("/", "send-meet-notification", users)
	})
}
