package handlers

import (
	"net/http"
	car_services_service "odo24_mobile_backend/api/services/car_services"
	"strconv"

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
	carID := c.MustGet("carID").(int64)

	services, err := ctrl.service.GetServices(carID, groupID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(services) == 0 {
		c.Status(http.StatusNoContent)
		c.Abort()
	} else {
		c.JSON(http.StatusOK, services)
	}
}

func (ctrl *CarServicesController) Create(c *gin.Context) {
	groupID := c.MustGet("groupID").(int64)
	carID := c.MustGet("carID").(int64)

	var body struct {
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
		CarID:        carID,
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

func (ctrl *CarServicesController) Update(c *gin.Context) {
	serviceID := c.MustGet("serviceID").(int64)

	var body struct {
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

	model := car_services_service.CarServiceUpdateModel{
		ServiceID:    serviceID,
		Odo:          body.Odo,
		NextDistance: body.NextDistance,
		Dt:           body.Dt,
		Description:  body.Description,
		Price:        body.Price,
	}
	err = ctrl.service.Update(model)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
	c.Abort()
}

func (ctrl *CarServicesController) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	serviceID := c.MustGet("serviceID").(int64)

	err := ctrl.service.Delete(userID, serviceID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
	c.Abort()
}

func (ctrl *CarServicesController) CheckParamServiceID(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	paramServiceID, ok := c.Params.Get("serviceID")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	serviceID, err := strconv.ParseInt(paramServiceID, 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err = ctrl.service.CheckOwner(userID, serviceID)
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Set("serviceID", serviceID)
}
