package handlers

import (
	"errors"
	"net/http"
	"odo24_mobile_backend/api/services"
	auth_service "odo24_mobile_backend/api/services/auth"
	"odo24_mobile_backend/config"
	"strings"

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
		if errors.Is(err, services.ErrorUnauthorize) {
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		return
	}

	c.JSON(http.StatusOK, token)
}

func (ctrl *AuthController) RefreshToken(c *gin.Context) {
	bearerToken := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(bearerToken, " ")
	if len(splitToken) < 2 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	accessToken := splitToken[1]

	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	result, err := ctrl.service.RefreshToken(accessToken, body.RefreshToken)
	if err != nil {
		if errors.Is(err, services.ErrorUnauthorize) {
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	var body struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}

	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = ctrl.service.ChangePassword(userID, body.CurrentPassword, body.NewPassword)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.String(http.StatusNoContent, "")
}
