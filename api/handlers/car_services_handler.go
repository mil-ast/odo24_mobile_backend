package handlers

import (
	"errors"
	"net/http"
	"odo24_mobile_backend/api/services"
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
	userID := c.MustGet("userID").(int64)

	var body struct {
		CarID int64 `form:"car_id" binding:"required"`
	}

	err := c.BindQuery(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	groupID, err := ctrl.getGroupIDFromParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	cars, err := ctrl.service.GetServices(userID, body.CarID, groupID)
	if err != nil {
		if errors.Is(err, services.ErrorNoPermission) {
			c.AbortWithStatus(http.StatusForbidden)
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, cars)
}

func (ctrl *CarServicesController) getGroupIDFromParams(c *gin.Context) (int64, error) {
	paramGroupID, ok := c.Params.Get("groupID")
	if !ok {
		return 0, errors.New("empty")
	}

	groupID, err := strconv.ParseInt(paramGroupID, 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return 0, errors.New("incorrect")
	}
	return groupID, nil
}
