package handlers

import (
	"log"
	"net/http"
	cars_service "odo24_mobile_backend/api/services/cars"
	groups_service "odo24_mobile_backend/api/services/groups"
	"odo24_mobile_backend/api/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CarsController struct {
	service       *cars_service.CarsService
	groupsService *groups_service.GroupsService
}

func NewCarsController(srv *cars_service.CarsService, groupsSrv *groups_service.GroupsService) *CarsController {
	return &CarsController{
		service:       srv,
		groupsService: groupsSrv,
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

	var carIDs []uint64

	for i := range cars {
		carIDs = append(carIDs, cars[i].CarID)
	}

	info, err := ctrl.service.GetCarNextServiceInformation(carIDs)
	if err != nil {
		log.Printf("getCarNextServiceInformation error: %v", err)
	} else {
		if info != nil {
			mapGroupIDs := make(map[uint64]struct{})
			var uniqGroupIDs []uint64
			for carID := range info {
				for groupID := range info[carID] {
					if _, ok := mapGroupIDs[groupID]; !ok {
						mapGroupIDs[groupID] = struct{}{}
						uniqGroupIDs = append(uniqGroupIDs, groupID)
					}
				}
			}

			groups, err := ctrl.groupsService.GetGroupsByIDs(uniqGroupIDs)
			if err != nil {
				log.Printf("GetGroupsByIDs error: %v", err)
			} else if len(groups) > 0 {
				mapGroups := make(map[uint64]groups_service.GroupModel)
				for i := range groups {
					mapGroups[groups[i].GroupID] = groups[i]
				}

				extInfo := make(map[uint64][]cars_service.CarExtData)
				for carID := range info {
					if _, ok := extInfo[carID]; !ok {
						extInfo[carID] = []cars_service.CarExtData{}
					}
					for groupID, data := range info[carID] {
						var groupName string

						if _, ok := mapGroups[groupID]; ok {
							groupName = mapGroups[groupID].Name
						} else {
							groupName = "Group " + strconv.FormatUint(groupID, 10)
						}

						extInfo[carID] = append(extInfo[carID], cars_service.CarExtData{
							Odo:       data.Odo,
							NextOdo:   data.NextOdo,
							GroupName: groupName,
						})
					}
				}

				for i := range cars {
					println(i)
					data, ok := extInfo[cars[i].CarID]
					if ok {
						for d := range data {
							next := data[d].Odo + data[d].NextOdo
							if next > cars[i].Odo {
								cars[i].CarExtData = append(cars[i].CarExtData, data[d])
							}
						}
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, cars)
}

func (ctrl *CarsController) Create(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)

	var body struct {
		Name   string `json:"name" binding:"required"`
		Odo    uint32 `json:"odo" binding:"required"`
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
		utils.BindServiceErrorWithAbort(c, "CartCreateError", "Не удалось создать авто", err)
		return
	}

	c.JSON(http.StatusOK, car)
}

func (ctrl *CarsController) Update(c *gin.Context) {
	carID := c.MustGet("carID").(uint64)

	var body struct {
		Name   string `json:"name" binding:"required"`
		Odo    uint32 `json:"odo" binding:"required"`
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
		utils.BindServiceErrorWithAbort(c, "CartUpdateError", "Не удалось изменить авто", err)
		return
	}

	utils.BindNoContent(c)
}

func (ctrl *CarsController) UpdateODO(c *gin.Context) {
	carID := c.MustGet("carID").(uint64)

	var body struct {
		Odo uint32 `json:"odo" binding:"required"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	err = ctrl.service.UpdateODO(carID, body.Odo)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "CartUpdateODOError", "Не удалось сохранить пробег авто", err)
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
		utils.BindErrorWithAbort(c, http.StatusForbidden, "forbidden", "Нет доступа", err)
		return
	}

	c.Set("carID", carID)
}
