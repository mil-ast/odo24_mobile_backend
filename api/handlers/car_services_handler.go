package handlers

import (
	"net/http"
	car_services_service "odo24_mobile_backend/api/services/car_services"

	"github.com/gin-gonic/gin"
)

type CarServicesController struct {
	service *car_services_service.CarServicesService
}

func NewCarServicesController() *CarServicesController {
	return &CarServicesController{
		service: car_services_service.NewCarServicesService(),
	}
}

func (ctrl *CarServicesController) GetGroupsByCurrentUser(c *gin.Context) {
	groupID := c.MustGet("groupID").(int64)

	var body struct {
		CarID int64 `form:"car_id" binding:"required"`
	}

	err := c.BindQuery(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	cars, err := ctrl.service.GetServices(body.CarID, groupID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, cars)
}

func (ctrl *CarServicesController) Create(c *gin.Context) {
	groupID := c.MustGet("groupID").(int64)

	var body struct {
		CarID        int64   `json:"car_id" binding:"required"`
		Odo          *uint32 `json:"odo" binding:"omitempty"`
		NextDistance *uint32 `json:"next_distance" binding:"omitempty"`
		Dt           string  `json:"dt" binding:"required"`
		Description  *string `json:"description" binding:"omitempty"`
		Price        *uint32 `json:"price" binding:"omitempty"`
	}
	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	model := car_services_service.CarServiceCreateModel{
		CarID:        body.CarID,
		GroupID:      groupID,
		Odo:          body.Odo,
		NextDistance: body.NextDistance,
		Dt:           body.Dt,
		Description:  body.Description,
		Price:        body.Price,
	}
	carService, err := ctrl.service.Create(model)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, carService)
}
