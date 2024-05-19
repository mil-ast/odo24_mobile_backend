package utils

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseError struct {
	Key     string `json:"key"`
	Message string `json:"message"`
}

func BindErrorWithAbort(c *gin.Context, statusCode int, key, message string, err error) {
	if err != nil {
		statusText := http.StatusText(statusCode)
		log.Printf("%s, err=%v", statusText, err)
	}

	c.JSON(statusCode, ResponseError{
		Key:     key,
		Message: message,
	})
	c.Abort()
}

func BindBadRequestWithAbort(c *gin.Context, message string, err error) {
	var errMessage = "Неверный запрос"

	if err != nil {
		log.Printf("BadRequest, err=%v", err)
	}

	if message != "" {
		errMessage = message
		statusText := http.StatusText(http.StatusBadRequest)
		log.Printf("%s, message=%s", statusText, message)
	}

	c.JSON(http.StatusBadRequest, ResponseError{
		Key:     "bad_request",
		Message: errMessage,
	})
	c.Abort()
}

func BindServiceErrorWithAbort(c *gin.Context, key, message string, err error) {
	BindErrorWithAbort(c, http.StatusInternalServerError, key, message, err)
}

func BindNoContent(c *gin.Context) {
	c.String(http.StatusNoContent, "")
}
