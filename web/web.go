package web

import (
	"fmt"

	socketio "github.com/googollee/go-socket.io"
)

// RegisterEvents registers web-related events
func RegisterEvents(server *socketio.Server) {
	server.OnEvent("/", "web_message", func(s socketio.Conn, msg string) {
		fmt.Println("Web Message:", msg)
		s.Emit("web_response", "Received: "+msg)
	})
}
