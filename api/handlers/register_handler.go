package handlers

import (
	"errors"
	"net/http"
	"net/mail"
	"time"

	register_service "odo24_mobile_backend/api/services/register"
	"odo24_mobile_backend/api/utils"
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
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	emailAddr, err := mail.ParseAddress(body.Email)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "Некорректный Email", err)
		return
	}

	err = ctrl.service.SendEmailCodeConfirmation(emailAddr)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "SendEmailCodeConfirmationError", "Не удалось отправить сообщение на почту", err)
		return
	}

	utils.BindNoContent(c)
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
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	emailAddr, err := mail.ParseAddress(body.Email)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "Некорректный Email", err)
		return
	}

	err = ctrl.service.RegisterByEmail(emailAddr, body.Code, body.Password)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			utils.BindErrorWithAbort(c, http.StatusForbidden, "ConfirmCodeError", "Неверный код подтверждения", err)
			return
		}
		if errors.Is(err, register_service.ErrLoginAlreadyExists) {
			utils.BindErrorWithAbort(c, http.StatusConflict, "LoginAlreadyExists", "Такой логин уже существует", err)
			return
		}
		utils.BindServiceErrorWithAbort(c, "RegisterByEmailError", "Непредвиденная ошибка регистрации", err)
		return
	}

	utils.BindNoContent(c)
}

func (ctrl *RegisterController) RecoverSendEmailCodeConfirmation(c *gin.Context) {
	time.Sleep(time.Second)

	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	err := c.Bind(&body)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	emailAddr, err := mail.ParseAddress(body.Email)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "Некорректный Email", err)
		return
	}

	err = ctrl.service.PasswordRecoverySendEmailCodeConfirmation(emailAddr)
	if err != nil {
		if errors.Is(err, register_service.ErrCodeHasAlreadyBeenSent) {
			utils.BindErrorWithAbort(c, http.StatusTooManyRequests, "CodeHasAlreadyBeenSent", "Код подтверждения уже был отправлен", err)
		} else {
			utils.BindServiceErrorWithAbort(c, "RecoverSendEmailCodeError", "Непредвиденная ошибка при отправке код подтверждения", err)
		}
		return
	}

	utils.BindNoContent(c)
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
		utils.BindBadRequestWithAbort(c, "", err)
		return
	}

	emailAddr, err := mail.ParseAddress(body.Email)
	if err != nil {
		utils.BindBadRequestWithAbort(c, "Некорректный Email", err)
		return
	}

	err = ctrl.service.PasswordRecovery(emailAddr, body.Code, body.Password)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			utils.BindErrorWithAbort(c, http.StatusForbidden, "ConfirmCodeError", "Неверный код подтверждения", err)
			return
		}
		utils.BindServiceErrorWithAbort(c, "PasswordRecoveryError", "Непредвиденная ошибка восстановления пароля", err)
		return
	}

	utils.BindNoContent(c)
}
