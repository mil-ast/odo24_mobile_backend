package handlers

import (
	"net/http"
	cars_service "odo24_mobile_backend/api/services/cars"
	"strconv"

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

func (ctrl *CarsController) Create(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	var body struct {
		Name   string `json:"name" binding:"required"`
		Odo    uint32 `json:"odo" binding:"required"`
		Avatar bool   `json:"avatar"`
	}
	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	model := cars_service.CarCreateModel{
		Name:   body.Name,
		Odo:    body.Odo,
		Avatar: body.Avatar,
	}
	car, err := ctrl.service.Create(userID, model)
	if err != nil {
		c.AbortWithError(http.StatusOK, err)
		return
	}

	c.JSON(http.StatusOK, car)
}

func (ctrl *CarsController) Update(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	carID := c.MustGet("carID").(int64)

	var body struct {
		Name   string `json:"name" binding:"required"`
		Odo    uint32 `json:"odo" binding:"required"`
		Avatar bool   `json:"avatar"`
	}
	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	model := cars_service.CarModel{
		CarID:  carID,
		Name:   body.Name,
		Odo:    body.Odo,
		Avatar: body.Avatar,
	}
	err = ctrl.service.Update(userID, model)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
	c.Abort()
}

func (ctrl *CarsController) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	carID := c.MustGet("carID").(int64)

	err := ctrl.service.Delete(userID, carID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
	c.Abort()
}

func (ctrl *CarsController) CheckParamCarID(c *gin.Context) {
	paramCarID, ok := c.Params.Get("carID")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	carID, err := strconv.ParseInt(paramCarID, 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	userID := c.MustGet("userID").(int64)

	err = ctrl.service.CheckOwner(carID, userID)
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Set("carID", carID)
}
