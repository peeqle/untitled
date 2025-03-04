package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"untitled/handle"
	"untitled/service"
	"untitled/service/auth"
)

var server = NewServer()

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections (customize for production)
	},
}

func main() {
	service.ConnectDB()
	service.ConnectMongo()

	r := gin.Default()
	handle.InitializeUserActivity(r)

	r.GET("/ws", auth.AuthMiddleware(), func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		server.HandleWebSocket(conn, c)
	})

	protectedGroup := r.Group("/api")
	protectedGroup.Use(auth.AuthMiddleware())

	runServer(r)
}

func runServer(engine *gin.Engine) {
	port := os.Getenv("SERVER_PORT")
	log.Println("Сервер запущен на http://localhost:" + port)
	engine.Run(":" + port)
}
