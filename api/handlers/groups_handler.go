package handlers

import (
	"net/http"
	groups_service "odo24_mobile_backend/api/services/groups"
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
	}
	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	model := groups_service.GroupCreateModel{
		Name: body.Name,
	}
	group, err := ctrl.service.Create(userID, model)
	if err != nil {
		c.AbortWithError(http.StatusOK, err)
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
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	model := groups_service.GroupModel{
		GroupID: groupID,
		Name:    body.Name,
	}
	err = ctrl.service.Update(userID, model)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
	c.Abort()
}

func (ctrl *GroupsController) UpdateSort(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	var body []int64

	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = ctrl.service.UpdateSort(userID, body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
	c.Abort()
}

func (ctrl *GroupsController) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	groupID := c.MustGet("groupID").(int64)

	err := ctrl.service.Delete(userID, groupID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
	c.Abort()
}

func (ctrl *GroupsController) CheckParamGroupID(c *gin.Context) {
	paramGroupID, ok := c.Params.Get("groupID")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	groupID, err := strconv.ParseInt(paramGroupID, 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	userID := c.MustGet("userID").(int64)

	err = ctrl.service.CheckOwner(groupID, userID)
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Set("groupID", groupID)
}
