package handlers

import (
	"net/http"
	groups_service "odo24_mobile_backend/api/services/groups"
	"odo24_mobile_backend/api/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GroupsController struct {
	service *groups_service.GroupsService
}

func NewGroupsController() *GroupsController {
	return &GroupsController{
		service: groups_service.NewGroupsService(),
	}
}

func (ctrl *GroupsController) GetGroupsByCurrentUser(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	groups, err := ctrl.service.GetGroupsByUser(userID)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "GetGroupsError", "Не удалось получить группы", err)
		return
	}

	if len(groups) == 0 {
		utils.BindNoContent(c)
	} else {
		c.JSON(http.StatusOK, groups)
	}
}

func (ctrl *GroupsController) Create(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	var body struct {
		Name string `json:"name" binding:"required"`
	}
	err := c.Bind(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	model := groups_service.GroupCreateModel{
		Name: body.Name,
	}
	group, err := ctrl.service.Create(userID, model)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "GroupsCreateError", "Не удалось создать группу", err)
		return
	}

	c.JSON(http.StatusOK, group)
}

func (ctrl *GroupsController) Update(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	groupID := c.MustGet("groupID").(int64)

	var body struct {
		Name string `json:"name" binding:"required"`
	}
	err := c.Bind(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	model := groups_service.GroupModel{
		GroupID: groupID,
		Name:    body.Name,
	}
	err = ctrl.service.Update(userID, model)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "GroupsUpdateError", "Не удалось изменить группу", err)
		return
	}

	utils.BindNoContent(c)
}

func (ctrl *GroupsController) UpdateSort(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	var body []int64

	err := c.Bind(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	err = ctrl.service.UpdateSort(userID, body)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "GroupsUpdateSortError", "Не удалось сохранить группировку групп", err)
		return
	}

	utils.BindNoContent(c)
}

func (ctrl *GroupsController) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	groupID := c.MustGet("groupID").(int64)

	err := ctrl.service.Delete(userID, groupID)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "GroupsDeleteError", "Не удалось удалить группу", err)
		return
	}

	utils.BindNoContent(c)
}

func (ctrl *GroupsController) CheckParamGroupID(c *gin.Context) {
	paramGroupID, ok := c.Params.Get("groupID")
	if !ok {
		utils.BindBadRequestWithAbort(c, "Параметр groupID обязателен", nil)
		return
	}

	groupID, err := strconv.ParseInt(paramGroupID, 10, 64)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "Ошибка парсинга группы", err)
		return
	}

	userID := c.MustGet("userID").(int64)

	err = ctrl.service.CheckOwner(groupID, userID)
	if err != nil {
		utils.BindErrorWithAbort(c, http.StatusForbidden, "forbidden", "Нет доступа", err)
		return
	}

	c.Set("groupID", groupID)
}
