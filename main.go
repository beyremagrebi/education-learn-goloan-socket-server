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

func main() {
	log.DEBUG = true
	c := socket.DefaultServerOptions()
	c.SetServeClient(true)
	// c.SetConnectionStateRecovery(&socket.ConnectionStateRecovery{})
	// c.SetAllowEIO3(true)
	c.SetPingInterval(300 * time.Millisecond)
	c.SetPingTimeout(200 * time.Millisecond)
	c.SetMaxHttpBufferSize(1000000)
	c.SetConnectTimeout(1000 * time.Millisecond)
	c.SetCors(&types.Cors{
		Origin:      "*",
		Credentials: true,
	})
	socketio := socket.NewServer(nil, nil)
	socketio.On("connection", func(clients ...interface{}) {
		client := clients[0].(*socket.Socket)
		client.On("newConnection", func(args ...interface{}) {
			fmt.Println("Connection")
			fmt.Println(args)
			client.Emit("newConnection", args...)
		})
		client.On("join", func(args ...interface{}) {
			room := args[0].(string)
			client.Join(socket.Room(room))
			client.Emit("joined", room)
		})
		client.On("typing", func(args ...interface{}) {
			room := args[0].(string)
			client.In(socket.Room(room)).Emit("typing", args...)
		})
		client.On("stop typing", func(args ...interface{}) {
			room := args[0].(string)
			client.In(socket.Room(room)).Emit("stop typing", args...)
		})
		client.On("accesDeniedForRule", func(args ...interface{}) {
			client.Emit("accesDeniedForRule", args...)
		})
		client.On("send-meet-notification", func(args ...interface{}) {
			client.Emit("send-meet-notification", args...)
		})
	})

	socketio.Of("/custom", nil).On("connection", func(clients ...interface{}) {
		client := clients[0].(*socket.Socket)
		client.Emit("auth", client.Handshake().Auth)
	})

	app := fiber.New()

	// app.Put("/socket.io", adaptor.HTTPHandler(socketio.ServeHandler(c))) // test
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
