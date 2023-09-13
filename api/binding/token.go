package binding

import (
	"net/http"
	auth_service "odo24_mobile_backend/api/services/auth"
	"strings"

	"github.com/gin-gonic/gin"
)

func Auth(c *gin.Context) {
	bearerToken := c.Request.Header.Get("Authorization")

	splitToken := strings.Split(bearerToken, " ")
	if len(splitToken) < 2 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	tokenClaims, err := auth_service.ValidateAccessToken(splitToken[1])
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims := *tokenClaims
	c.Set("userID", int64(claims["uid"].(float64)))

	c.Next()
}
