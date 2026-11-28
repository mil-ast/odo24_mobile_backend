package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	app_version_service "github.com/mil-ast/odo24_mobile_backend/api/services/app_version"
	"github.com/mil-ast/odo24_mobile_backend/api/utils"
)

type AppVersionController struct {
	service *app_version_service.AppVersionService
}

func NewAppVersionController(srv *app_version_service.AppVersionService) *AppVersionController {
	return &AppVersionController{
		service: srv,
	}
}

func (ctrl *AppVersionController) Get(c *gin.Context) {
	version, err := ctrl.service.GetCurrentVersion()
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "GetAppVersionError", "Не удалось получить версию приложения", err)
		return
	}

	c.JSON(http.StatusOK, version)
}
