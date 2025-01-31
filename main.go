package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	socketio "github.com/googollee/go-socket.io"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

type User struct {
	UserID   string `json:"userId"`
	SocketID string `json:"socketId"`
	Status   string `json:"status"`
}

var connectedUsers = sync.Map{}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Create a new Socket.IO server
	server := socketio.NewServer(nil)

	// Handle connections
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Println("connected:", s.ID())
		return nil
	})

	// Handle "newConnection" event
	server.OnEvent("/", "newConnection", func(s socketio.Conn, userId string) {
		log.Println("newConnection:", userId)
		connectedUsers.Store(userId, User{UserID: userId, SocketID: s.ID(), Status: "online"})
		broadcastConnectedUsers(server)
	})

	// Handle "setup" event
	server.OnEvent("/", "setup", func(s socketio.Conn, userData map[string]interface{}) {
		userId := userData["_id"].(string)
		firstName := userData["firstName"].(string)
		lastName := userData["lastName"].(string)
		log.Println(firstName, lastName, "connected")
		connectedUsers.Store(userId, User{UserID: userId, SocketID: s.ID(), Status: "online"})
		broadcastConnectedUsers(server)
	})

	// Handle "join" event
	server.OnEvent("/", "join", func(s socketio.Conn, room string) {
		log.Println("User Joined Room:", room)
		s.Join(room)
		s.Emit("joined", map[string]string{"room": room})
	})

	// Handle "typing" event
	server.OnEvent("/", "typing", func(s socketio.Conn, room string) {
		server.BroadcastToRoom("/", room, "typing", nil)
	})

	// Handle "stop typing" event
	server.OnEvent("/", "stop typing", func(s socketio.Conn, room string) {
		server.BroadcastToRoom("/", room, "stop typing", nil)
	})

	// Handle "new message" event
	server.OnEvent("/", "new message", func(s socketio.Conn, newMessageReceived string) {
		var messageData []interface{}
		err := json.Unmarshal([]byte(newMessageReceived), &messageData)
		if err != nil {
			log.Println("Error decoding new message:", err)
			return
		}
		handleNewMessage(server, messageData)
	})

	// Handle "disconnect" event
	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		deleteUser(s)
		broadcastConnectedUsers(server)
		log.Println("disconnected:", s.ID(), reason)
	})

	// Serve static files (optional)
	http.Handle("/socket.io/", server)
	// Add CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow your frontend origin
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	// Wrap the Socket.IO handler with CORS
	handler := c.Handler(http.DefaultServeMux)
	http.Handle("/", http.FileServer(http.Dir("./public")))

	// Start the server
	log.Println("Socket.IO server started on :8800")
	if err := http.ListenAndServe(":8800", handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// Broadcast connected users to all clients
func broadcastConnectedUsers(server *socketio.Server) {
	users := []User{}
	connectedUsers.Range(func(key, value interface{}) bool {
		users = append(users, value.(User))
		return true
	})
	server.BroadcastToNamespace("/", "connectedUsers", users)
}

// Handle new messages
func handleNewMessage(server *socketio.Server, messageData []interface{}) {
	// Implement your logic for handling new messages
}

// Delete a user from the connectedUsers map
func deleteUser(s socketio.Conn) {
	connectedUsers.Range(func(key, value interface{}) bool {
		if value.(User).SocketID == s.ID() {
			connectedUsers.Delete(key)
			return false
		}
		return true
	})
}
