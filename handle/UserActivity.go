package handle

import (
	"github.com/gin-gonic/gin"
	"untitled/service/auth"
)

func InitializeUserActivity(engine *gin.Engine) {
	engine.POST("/register", auth.Register)
}
