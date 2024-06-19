package handlers

import (
	"net/http"
	car_services_service "odo24_mobile_backend/api/services/car_services"
	"odo24_mobile_backend/api/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CarServicesController struct {
	service *car_services_service.CarServicesService
}

func NewCarServicesController(srv *car_services_service.CarServicesService) *CarServicesController {
	return &CarServicesController{
		service: srv,
	}
}

func (ctrl *CarServicesController) GetServicesByCurrentUserAndGroup(c *gin.Context) {
	groupID := c.MustGet("groupID").(int64)
	carID := c.MustGet("carID").(int64)

	services, err := ctrl.service.GetServices(carID, groupID)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "GetServices", "Не удалось получить список записей", err)
		return
	}

	if len(services) == 0 {
		utils.BindNoContent(c)
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
	err := c.ShouldBindJSON(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", nil)
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
		utils.BindServiceErrorWithAbort(c, "ServiceCreateError", "Не удалось создать запись", err)
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
	err := c.ShouldBindJSON(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", nil)
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
		utils.BindServiceErrorWithAbort(c, "ServiceUpdateError", "Не удалось обновить запись", err)
		return
	}

	utils.BindNoContent(c)
}

func (ctrl *CarServicesController) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)
	serviceID := c.MustGet("serviceID").(int64)

	err := ctrl.service.Delete(userID, serviceID)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "ServiceDeleteError", "Не удалось удалить запись", err)
		return
	}

	utils.BindNoContent(c)
}

func (ctrl *CarServicesController) CheckParamServiceID(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)
	paramServiceID, ok := c.Params.Get("serviceID")
	if !ok {
		utils.BindBadRequestWithAbort(c, "Параметр serviceID обязателен", nil)
		return
	}

	serviceID, err := strconv.ParseUint(paramServiceID, 10, 64)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "Ошибка парсинга serviceID", err)
		return
	}

	err = ctrl.service.CheckOwner(userID, serviceID)
	if err != nil {
		utils.BindErrorWithAbort(c, http.StatusForbidden, "forbidden", "Нет доступа", err)
		return
	}

	c.Set("serviceID", serviceID)
}
