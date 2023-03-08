package handlers

import (
	"net/http"
	auth_service "odo24_mobile_backend/api/services/auth"
	"odo24_mobile_backend/config"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service *auth_service.AuthService
}

func NewAuthController() *AuthController {
	cfg := config.GetInstance().App
	return &AuthController{
		service: auth_service.NewAuthService(cfg.JwtAccessSecret, cfg.JwtRefreshSecret, cfg.PasswordSalt),
	}
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"login" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	token, err := ctrl.service.Login(body.Email, body.Password)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, token)
}
