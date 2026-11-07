package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mil-ast/odo24_mobile_backend/api/services"
	cars_service "github.com/mil-ast/odo24_mobile_backend/api/services/cars"
	"github.com/mil-ast/odo24_mobile_backend/api/utils"
)

type CarsController struct {
	service *cars_service.CarsService
}

func NewCarsController(srv *cars_service.CarsService) *CarsController {
	return &CarsController{
		service: srv,
	}
}

func (ctrl *CarsController) GetCarsByCurrentUser(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)
	cars, err := ctrl.service.GetCarsByUser(userID)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "GetCarsError", "Не удалось получить авто", err)
		return
	}

	if len(cars) == 0 {
		utils.BindNoContent(c)
		return
	}

	c.JSON(http.StatusOK, cars)
}

func (ctrl *CarsController) Create(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)

	var body struct {
		Name   string `json:"name" binding:"required"`
		Odo    uint32 `json:"odo"`
		Avatar bool   `json:"avatar"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	model := cars_service.CarCreateModel{
		Name:   body.Name,
		Odo:    body.Odo,
		Avatar: body.Avatar,
	}
	car, err := ctrl.service.Create(userID, model)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "CarCreateError", "Не удалось создать авто", err)
		return
	}

	c.JSON(http.StatusOK, car)
}

func (ctrl *CarsController) Update(c *gin.Context) {
	carID := c.MustGet("carID").(uint64)

	var body struct {
		Name   string `json:"name" binding:"required"`
		Odo    uint32 `json:"odo"`
		Avatar bool   `json:"avatar"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	model := cars_service.CarModel{
		CarID:  carID,
		Name:   body.Name,
		Odo:    body.Odo,
		Avatar: body.Avatar,
	}
	err = ctrl.service.Update(model)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "CarUpdateError", "Не удалось изменить авто", err)
		return
	}

	utils.BindNoContent(c)
}

func (ctrl *CarsController) UpdateODO(c *gin.Context) {
	carID := c.MustGet("carID").(uint64)

	var body struct {
		Odo uint32 `json:"odo"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	err = ctrl.service.UpdateODO(carID, body.Odo)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "CarUpdateODOError", "Не удалось сохранить пробег авто", err)
		return
	}

	utils.BindNoContent(c)
}

func (ctrl *CarsController) Delete(c *gin.Context) {
	carID := c.MustGet("carID").(uint64)

	err := ctrl.service.Delete(carID)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "CarDeleteError", "Не удалось удалить авто", err)
		return
	}

	utils.BindNoContent(c)
}

func (ctrl *CarsController) CheckParamCarID(c *gin.Context) {
	paramCarID, ok := c.Params.Get("carID")
	if !ok {
		utils.BindBadRequestWithAbort(c, "Параметр carID обязателен", nil)
		return
	}

	carID, err := strconv.ParseUint(paramCarID, 10, 64)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "Ошибка парсинга carID", err)
		return
	}

	userID := c.MustGet("userID").(uint64)

	err = ctrl.service.CheckOwner(carID, userID)
	if err != nil {
		if errors.Is(err, services.ErrorNoPermission) {
			utils.BindErrorWithAbort(c, http.StatusForbidden, "forbidden", "Нет доступа", err)
			return
		}
		utils.BindServiceErrorWithAbort(c, "CarOwnerCheckError", "Не удалось проверить доступ", err)
		return
	}

	c.Set("carID", carID)
}
