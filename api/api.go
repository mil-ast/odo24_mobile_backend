package api

import (
	"odo24_mobile_backend/api/handlers"

	"github.com/gin-gonic/gin"
)

func InitHandlers() *gin.Engine {
	r := gin.Default()

	r.GET("/api/ping", handlers.Ping)

	registerCtrl := handlers.NewRegisterController()
	apiRegister := r.Group("/api/register")
	apiRegister.POST("/send_code", registerCtrl.SendEmailCodeConfirmation)
	apiRegister.POST("/register_by_email", registerCtrl.RegisterByEmail)

	return r
}
