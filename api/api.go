package api

import (
	"odo24_mobile_backend/api/handlers"

	"github.com/gin-gonic/gin"
)

func InitHandlers() *gin.Engine {
	r := gin.Default()

	r.GET("/api/ping", handlers.Ping)

	//register
	registerCtrl := handlers.NewRegisterController()
	apiRegister := r.Group("/api/register")
	apiRegister.POST("/register_send_code", registerCtrl.SendEmailCodeConfirmation)
	apiRegister.POST("/register_by_email", registerCtrl.RegisterByEmail)
	apiRegister.POST("/recover_send_code", registerCtrl.RecoverSendEmailCodeConfirmation)
	apiRegister.POST("/recover_password", registerCtrl.RecoverPassword)

	//auth
	authCtrl := handlers.NewAuthController()

	apiAuth := r.Group("/api/auth")
	apiAuth.POST("/login", authCtrl.Login)
	apiAuth.POST("/refresh_token", authCtrl.RefreshToken)
	apiAuth.POST("/change_password", authCtrl.CheckAuth, authCtrl.ChangePassword)

	//cars
	carsCtrl := handlers.NewCarsController()
	apiCars := r.Group("/api/cars", authCtrl.CheckAuth)
	apiCars.GET("", carsCtrl.GetCarsByCurrentUser)
	apiCars.POST("", carsCtrl.Create)

	apiCarsID := apiCars.Group("/:carID", carsCtrl.CheckParamCarID)
	apiCarsID.PUT("", carsCtrl.Update)
	apiCarsID.PUT("/update_odo", carsCtrl.UpdateODO)
	apiCarsID.DELETE("", carsCtrl.Delete)

	//groups
	groupsCtrl := handlers.NewGroupsController()
	apiGroups := r.Group("/api/groups", authCtrl.CheckAuth)
	apiGroups.GET("", groupsCtrl.GetGroupsByCurrentUser)
	apiGroups.POST("", groupsCtrl.Create)
	apiGroups.POST("/update_sort", groupsCtrl.UpdateSort)
	apiGroupsID := apiGroups.Group("/:groupID", groupsCtrl.CheckParamGroupID)
	apiGroupsID.PUT("", groupsCtrl.Update)
	apiGroupsID.DELETE("", groupsCtrl.Delete)

	//car services
	apiServiceCtrl := apiCarsID.Group("/groups/:groupID/services", groupsCtrl.CheckParamGroupID)
	carServicesCtrl := handlers.NewCarServicesController()
	apiServiceCtrl.GET("", carServicesCtrl.GetServicesByCurrentUserAndGroup)
	apiServiceCtrl.POST("", carServicesCtrl.Create)
	apiServiceCtrlID := r.Group("/api/services/:serviceID", authCtrl.CheckAuth, carServicesCtrl.CheckParamServiceID)
	apiServiceCtrlID.PUT("", carServicesCtrl.Update)
	apiServiceCtrlID.DELETE("", carServicesCtrl.Delete)

	return r
}
