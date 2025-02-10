package utils

import "github.com/zishang520/socket.io/v2/socket"

func RoomJoined(client *socket.Socket, room string) {
	client.Emit("joined", room)
}
