package api

import (
	"odo24_mobile_backend/api/binding"
	"odo24_mobile_backend/api/handlers"

	"github.com/gin-gonic/gin"
)

func InitHandlers() *gin.Engine {
	r := gin.Default()

	r.GET("/api/ping", handlers.Ping)

	// register
	registerCtrl := handlers.NewRegisterController()
	apiRegister := r.Group("/api/register")
	apiRegister.POST("/send_code", registerCtrl.SendEmailCodeConfirmation)
	apiRegister.POST("/register_by_email", registerCtrl.RegisterByEmail)

	//auth
	authCtrl := handlers.NewAuthController()
	apiAuth := r.Group("/api/auth")
	apiAuth.POST("/login", authCtrl.Login)
	apiAuth.POST("/refresh_token", authCtrl.RefreshToken)
	apiAuth.POST("/change_password", binding.Auth, authCtrl.ChangePassword)

	//cars
	carsCtrl := handlers.NewCarsController()
	apiCars := r.Group("/api/cars", binding.Auth)
	apiCars.GET("/get", carsCtrl.GetCarsByCurrentUser)

	return r
}
