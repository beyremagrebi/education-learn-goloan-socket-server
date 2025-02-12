package web

import (
	"fmt"

	socketio "github.com/googollee/go-socket.io"
)

func RegisterEvents(server *socketio.Server) {
	server.OnEvent("/", "join", func(s socketio.Conn, room string) {
		fmt.Printf("User joined to room %s\n", room)
		s.Join(room)
	})

	server.OnEvent("/", "send-meet-notification", func(s socketio.Conn, users any) {
		server.BroadcastToRoom("/", "studiffy", "send-meet-notification", users)
	})
	server.OnEvent("/", "accesDeniedForRule", func(s socketio.Conn, data interface{}) {
		server.BroadcastToRoom("/", "studiffy", "accesDeniedForRule", data)
	})
}
