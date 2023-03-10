package handlers

import (
	"net/http"
	groups_service "odo24_mobile_backend/api/services/groups"

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
	cars, err := ctrl.service.GetGroupsByUser(userID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, cars)
}

func (ctrl *GroupsController) Create(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	var body struct {
		Name string `json:"name" binding:"required"`
		Sort uint32 `json:"sort" binding:"required"`
	}
	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	model := groups_service.GroupCreateModel{
		Name: body.Name,
		Sort: body.Sort,
	}
	group, err := ctrl.service.Create(userID, model)
	if err != nil {
		c.AbortWithError(http.StatusOK, err)
		return
	}

	c.JSON(http.StatusOK, group)
}
