package handlers

import (
	"errors"
	"net/http"
	"net/mail"
	"time"

	register_service "odo24_mobile_backend/api/services/register"
	"odo24_mobile_backend/config"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-gonic/gin"
)

type RegisterController struct {
	service *register_service.RegisterService
}

func NewRegisterController() *RegisterController {
	cfg := config.GetInstance().App
	return &RegisterController{
		service: register_service.NewRegisterService(cfg.PasswordSalt),
	}
}

func (ctrl *RegisterController) SendEmailCodeConfirmation(c *gin.Context) {
	time.Sleep(time.Second)

	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	emailAddr, err := mail.ParseAddress(body.Email)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = ctrl.service.SendEmailCodeConfirmation(emailAddr)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusNoContent, "")
}

func (ctrl *RegisterController) RegisterByEmail(c *gin.Context) {
	time.Sleep(time.Second)

	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Code     uint16 `json:"code" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	emailAddr, err := mail.ParseAddress(body.Email)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = ctrl.service.RegisterByEmail(emailAddr, body.Code, body.Password)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) || errors.Is(err, register_service.ErrLoginAlreadyExists) {
			c.AbortWithError(http.StatusForbidden, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, http.StatusText(http.StatusOK))
}

func (ctrl *RegisterController) RecoverSendEmailCodeConfirmation(c *gin.Context) {
	time.Sleep(time.Second)

	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	emailAddr, err := mail.ParseAddress(body.Email)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = ctrl.service.PasswordRecoverySendEmailCodeConfirmation(emailAddr)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusNoContent, "")
}

func (ctrl *RegisterController) RecoverPassword(c *gin.Context) {
	time.Sleep(time.Second * 3)

	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Code     uint16 `json:"code" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	err := c.Bind(&body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	emailAddr, err := mail.ParseAddress(body.Email)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = ctrl.service.PasswordRecovery(emailAddr, body.Code, body.Password)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			c.AbortWithError(http.StatusForbidden, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, http.StatusText(http.StatusOK))
}
