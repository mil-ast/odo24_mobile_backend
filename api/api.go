package api

import (
	"odo24_mobile_backend/api/binding"
	"odo24_mobile_backend/api/handlers"

	"github.com/gin-gonic/gin"
)

func InitHandlers() *gin.Engine {
	r := gin.Default()

	r.GET("/api/ping", handlers.Ping)

	//register
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
	apiCars.GET("", carsCtrl.GetCarsByCurrentUser)
	apiCars.POST("", carsCtrl.Create)
	apiCars.PUT("/:carID", carsCtrl.CheckParamCarID, carsCtrl.Update)
	apiCars.DELETE("/:carID", carsCtrl.CheckParamCarID, carsCtrl.Delete)

	//groups
	groupsCtrl := handlers.NewGroupsController()
	apiGroups := r.Group("/api/groups", binding.Auth)
	apiGroups.GET("", groupsCtrl.GetGroupsByCurrentUser)
	apiGroups.POST("", groupsCtrl.Create)
	apiGroups.PUT("/:groupID", groupsCtrl.CheckParamGroupID, groupsCtrl.Update)
	apiGroups.DELETE("/:groupID", groupsCtrl.CheckParamGroupID, groupsCtrl.Delete)

	carServicesCtrl := handlers.NewCarServicesController()
	apiGroups.GET("/:groupID/services", groupsCtrl.CheckParamGroupID, carServicesCtrl.GetGroupsByCurrentUser)
	apiGroups.POST("/:groupID/services", groupsCtrl.CheckParamGroupID, carServicesCtrl.Create)

	return r
}
