package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/zishang520/engine.io/v2/log"
	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/socket.io/v2/socket"
)

var connectedUsers = make(map[string]map[string]string)

type Permission struct {
	Create bool `mapstructure:"create"`
	Read   bool `mapstructure:"read"`
	Update bool `mapstructure:"update"`
	Delete bool `mapstructure:"delete"`
	Export bool `mapstructure:"export"`
	Import bool `mapstructure:"import"`
}

func main() {
	log.DEBUG = true
	c := socket.DefaultServerOptions()
	c.SetServeClient(true)
	c.SetPingInterval(60 * time.Second)
	c.SetPingTimeout(60 * time.Second)
	c.SetMaxHttpBufferSize(1000000)
	c.SetConnectTimeout(1000 * time.Millisecond)
	c.SetCors(&types.Cors{
		Origin:      "*",
		Credentials: true,
	})

	socketio := socket.NewServer(nil, nil)

	socketio.On("connection", func(clients ...interface{}) {
		client := clients[0].(*socket.Socket)
		fmt.Println("New client connected")

		client.On("newConnection", func(args ...interface{}) {
			userId := args[0].(string)
			fmt.Println("User connected:", userId)
			connectedUsers[userId] = map[string]string{
				"userId":   userId,
				"socketId": string(client.Id()),
				"status":   "online",
			}
			client.Emit("connectedUsers", connectedUsers)
		})

		client.On("setup", func(args ...interface{}) {
			userData := args[0].(map[string]interface{})
			userId := userData["_id"].(string)
			fmt.Println("User setup:", userId)
			connectedUsers[userId] = map[string]string{
				"userId":   userId,
				"socketId": string(client.Id()),
				"status":   "online",
			}
			client.Emit("connectedUsers", connectedUsers)
		})

		client.On("join", func(args ...interface{}) {
			room := args[0].(string)
			roomObj := socket.Room(room)
			client.Join(roomObj)
			fmt.Println("User joined room:", room)
			client.Emit("joined", map[string]string{"room": room})
		})

		client.On("typing", func(args ...interface{}) {
			room := args[0].(string)
			roomObj := socket.Room(room)
			client.To(roomObj).Emit("typing")
		})

		client.On("stop typing", func(args ...interface{}) {
			room := args[0].(string)
			roomObj := socket.Room(room)
			client.To(roomObj).Emit("stop typing")
		})

		client.On("new message", func(args ...interface{}) {
			messageData := args[0].(map[string]interface{})
			chatUsers := messageData["users"].([]interface{})
			fmt.Println("New message received:", messageData)
			for _, user := range chatUsers {
				userId := user.(map[string]interface{})["_id"].(string)
				if receiver, exists := connectedUsers[userId]; exists {
					fmt.Println(receiver)
					client.In(socket.Room(userId)).Emit("message received", messageData)
					fmt.Println("Message sent to:", userId)
				}
			}
		})
		client.On("accesDeniedForRule", func(args ...interface{}) {
			fmt.Println(connectedUsers)
			if len(args) < 1 {
				fmt.Println("Invalid arguments for accesDeniedForRule event")
				return
			}

			// Extract rule (permissions object) from the map
			rawRule, ok := args[0].(map[string]interface{})
			if !ok {
				fmt.Println("Failed to parse rule")
				return
			}

			// Extract userId from the rule map
			userId, ok := rawRule["user"].(string)
			if !ok {
				fmt.Println("Failed to extract user ID from rule")
				return
			}

			// Remove the "user" field from the rule so it's separate
			delete(rawRule, "user")

			fmt.Println("Access denied for rule:", rawRule, "User ID:", userId)

			// Emit with the correct structure
			client.Broadcast().Emit("accesDeniedForRule", map[string]interface{}{
				"rule":   rawRule, // Rule object without userId
				"userId": userId,  // Separate userId field
			})
		})
		client.On("disconnect", func(args ...interface{}) {
			for userId, user := range connectedUsers {
				if user["socketId"] == string(client.Id()) {
					user["status"] = "offline"
					connectedUsers[userId] = user
					client.Leave(socket.Room(userId))
					client.Emit("connectedUsers", connectedUsers)
					fmt.Println(userId, "disconnected")
					break
				}
			}
		})
	})

	app := fiber.New()
	app.Get("/socket.io", adaptor.HTTPHandler(socketio.ServeHandler(c)))
	app.Post("/socket.io", adaptor.HTTPHandler(socketio.ServeHandler(c)))
	go app.Listen(":8800")

	exit := make(chan struct{})
	SignalC := make(chan os.Signal, 1)

	signal.Notify(SignalC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range SignalC {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				close(exit)
				return
			}
		}
	}()

	<-exit
	socketio.Close(nil)
	os.Exit(0)
}
