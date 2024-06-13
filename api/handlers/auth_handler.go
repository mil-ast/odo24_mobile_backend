package handlers

import (
	"errors"
	"log"
	"net/http"
	"odo24_mobile_backend/api/services"
	auth_service "odo24_mobile_backend/api/services/auth"
	"odo24_mobile_backend/api/utils"
	"odo24_mobile_backend/config"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthController struct {
	service *auth_service.AuthService
}

func NewAuthController() *AuthController {
	cfg := config.GetInstance().App
	return &AuthController{
		service: auth_service.NewAuthService(cfg.JwtAccessPrivateKeyPath, cfg.JwtAccessPublicKeyPath, cfg.JwtRefreshPrivateKeyPath, cfg.JwtRefreshPublicKeyPath),
	}
}

// проверка авторизации в апи
func (ctrl *AuthController) CheckAuth(c *gin.Context) {
	bearerToken := c.Request.Header.Get("Authorization")

	splitToken := strings.Split(bearerToken, " ")
	if len(splitToken) < 2 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := ctrl.service.ParseAccessToken(splitToken[1])
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			log.Println("check token error: ErrTokenExpired")
		} else {
			log.Printf("check token error: %v\n", err)
		}
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("check token error: Claims is empty")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set("userID", int64(claims["uid"].(float64)))
	c.Next()
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"login" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	token, err := ctrl.service.Login(body.Email, body.Password)
	if err != nil {
		if errors.Is(err, services.ErrorUnauthorize) {
			utils.BindErrorWithAbort(c, http.StatusUnauthorized, "AuthError", "Неверный логин или пароль", nil)
		} else {
			utils.BindServiceErrorWithAbort(c, "LoginError", "Произошла ошибка при авторизации", err)
		}
		return
	}

	c.JSON(http.StatusOK, token)
}

func (ctrl *AuthController) RefreshToken(c *gin.Context) {
	bearerToken := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(bearerToken, " ")
	if len(splitToken) < 2 {
		utils.BindErrorWithAbort(c, http.StatusUnauthorized, "RefreshTokenError", "Некорректный токен авторизации", nil)
		return
	}

	accessToken := splitToken[1]

	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	result, err := ctrl.service.RefreshToken(accessToken, body.RefreshToken)
	if err != nil {
		if errors.Is(err, services.ErrorUnauthorize) {
			utils.BindErrorWithAbort(c, http.StatusUnauthorized, "RefreshError", "Ошибка обновления токена. Попробуйте переавторизоваться", err)
		} else {
			utils.BindServiceErrorWithAbort(c, "RefreshTokenError", "Ошибка обновления токена", err)
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

	err := c.ShouldBindJSON(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	err = ctrl.service.ChangePassword(userID, body.CurrentPassword, body.NewPassword)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "ChangePasswordError", "Ошибка изменения пароля", err)
		return
	}

	utils.BindNoContent(c)
}
