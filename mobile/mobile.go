package mobile

import (
	"fmt"

	socketio "github.com/googollee/go-socket.io"
)

// RegisterEvents registers mobile-related events
func RegisterEvents(server *socketio.Server) {
	server.OnEvent("/", "mobile_message", func(s socketio.Conn, msg string) {
		fmt.Println("Mobile Message:", msg)
		s.Emit("mobile_response", "Received: "+msg)

	})
}
