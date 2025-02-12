package mobile

import (
	"fmt"

	"github.com/fatih/color"
	socketio "github.com/googollee/go-socket.io"
	"github.com/proservices/socket-golang-server/types"
)

var ConnectedUsersByFacilities = make(map[string][]types.User)

func RegisterEvents(server *socketio.Server) {

	server.OnEvent("/", "join-facility", func(userSocket socketio.Conn, facilityId string) {
		userSocket.Join(facilityId)
	})

	server.OnEvent("/", "new-connection-by-facilities", func(userSocket socketio.Conn, data types.UserConnection) {

		var users = []types.User{}

		if existingUsers, ok := ConnectedUsersByFacilities[data.FacilityId]; ok {
			users = existingUsers
		} else {
			users = []types.User{}
		}

		userAlreadyInRoom := false
		for _, user := range users {
			if user.UserID == data.UserId {
				userAlreadyInRoom = true
				break
			}
		}
		if !userAlreadyInRoom {
			userSocket.Join(data.UserId)
			userSocket.Emit("/", "room-created", data.UserId)

			users = append(users, types.User{
				UserID:   data.UserId,
				SocketID: userSocket.ID(),
				Status:   "online",
			})
			fmt.Printf("%s was connected => {\n user-id : %s \n facility-id: %s\n}\n",
				color.BlueString(data.FullName),
				color.GreenString(data.UserId),
				color.GreenString(data.FacilityId))
			ConnectedUsersByFacilities[data.FacilityId] = users

		}

		server.BroadcastToRoom("/", data.FacilityId, "new-user-connected", ConnectedUsersByFacilities[data.FacilityId])
	})

	server.OnEvent("/", "disconnect-by-facilities", func(userSocket socketio.Conn, data types.UserConnection) {
		if users, ok := ConnectedUsersByFacilities[data.FacilityId]; ok {
			var updatedUsers []types.User
			var disconnectedUser types.User
			userFound := false

			for _, user := range users {
				if user.UserID == data.UserId {
					disconnectedUser = user
					userFound = true
					continue
				}
				updatedUsers = append(updatedUsers, user)
			}

			if userFound {
				userSocket.LeaveAll()

				fmt.Printf("%s was disconnected => {\n user-id : %s \n facility-id: %s\n}\n",
					color.RedString(data.FullName),
					color.YellowString(data.UserId),
					color.YellowString(data.FacilityId))

				ConnectedUsersByFacilities[data.FacilityId] = updatedUsers

				server.BroadcastToRoom("/", data.FacilityId, "disconnect-user", disconnectedUser)

				if len(updatedUsers) == 0 {
					delete(ConnectedUsersByFacilities, data.FacilityId)
				}
			}
		}
	})

	server.OnEvent("/", "join-chatroom", func(userSocket socketio.Conn, room string) {
		rooms := userSocket.Rooms()

		inRoom := false
		for _, r := range rooms {
			if r == room {
				inRoom = true
				break
			}
		}

		if !inRoom {
			userSocket.Join(room)
			userSocket.Emit("joined-chatroom", map[string]string{"room": room})
		}
	})
	server.OnEvent("/", "send-message-mobile", func(userSocket socketio.Conn, message string, messageId string, chatId string, userId string, senderId string) {
		data := map[string]interface{}{
			"userId":    userId,
			"message":   message,
			"messageId": messageId,
			"chatId":    chatId,
			"senderId":  senderId,
		}

		server.BroadcastToRoom("/", chatId, "message-recieved-mobile", data)
	})

	server.OnEvent("/", "read-message", func(userSocket socketio.Conn, room string, userId string, messageId string) {
		messageData := map[string]interface{}{
			"userId":    userId,
			"room":      room,
			"messageId": messageId,
		}
		server.BroadcastToRoom("/", room, "message-readed", messageData)
	})
	server.OnEvent("/", "typing-mobile", func(userSocket socketio.Conn, room string, userId string) {
		typingData := map[string]interface{}{
			"userId": userId,
			"room":   room,
		}
		server.BroadcastToRoom("/", room, "typing-mobile", typingData)
	})
	server.OnEvent("/", "stop-typing-mobile", func(userSocket socketio.Conn, room string, userId string) {
		stopTypingData := map[string]interface{}{
			"room":   room,
			"userId": userId,
		}
		server.BroadcastToRoom("/", room, "stop-typing-mobile", stopTypingData)
	})

	server.OnEvent("/", "check-private-room", func(userSocket socketio.Conn, chatId string, userId string) {
		updateRoomData := map[string]interface{}{
			"chatId": chatId,
		}
		server.BroadcastToRoom("/", userId, "update-private-room", updateRoomData)
	})
	server.OnEvent("/", "check-group-room", func(userSocket socketio.Conn, chatId string) {
		updateRoomData := map[string]interface{}{
			"chatId": chatId,
		}
		server.BroadcastToRoom("/", chatId, "update-group-room", updateRoomData)
	})

}
