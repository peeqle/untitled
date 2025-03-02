package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"untitled/handle"
	"untitled/service"
)

func main() {
	service.ConnectDB()

	r := gin.Default()
	handle.InitializeUserActivity(r)

	runServer(r)
}

func runServer(engine *gin.Engine) {
	port := os.Getenv("SERVER_PORT")
	log.Println("Сервер запущен на http://localhost:" + port)
	engine.Run(":" + port)
}
