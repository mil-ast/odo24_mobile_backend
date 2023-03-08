package handlers

import (
	"net/http"
	cars_service "odo24_mobile_backend/api/services/cars"

	"github.com/gin-gonic/gin"
)

type CarsController struct {
	service *cars_service.CarsService
}

func NewCarsController() *CarsController {
	return &CarsController{
		service: cars_service.NewCarsService(),
	}
}

func (ctrl *CarsController) GetCarsByCurrentUser(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	cars, err := ctrl.service.GetCarsByUser(userID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, cars)
}
