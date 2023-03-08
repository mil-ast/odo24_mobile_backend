package auth_service

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"odo24_mobile_backend/config"
	"odo24_mobile_backend/db"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var errorUnauthorize = errors.New("unauthorize")

type AuthService struct {
	jwtTokenSecret   string
	jwtRefreshSecret string
	passwordSalt     string
}

func NewAuthService(jwtTokenSecret, jwtRefreshSecret, passwordSalt string) *AuthService {
	return &AuthService{
		jwtTokenSecret:   jwtTokenSecret,
		jwtRefreshSecret: jwtRefreshSecret,
		passwordSalt:     passwordSalt,
	}
}

func (srv *AuthService) Login(email string, password string) (*AuthResultModel, error) {
	pg := db.Conn()
	var user struct {
		UserID   int64
		Password []byte
	}
	err := pg.QueryRow("select u.user_id,u.password_hash from profiles.users u where u.login = $1", email).Scan(&user.UserID, &user.Password)
	if err != nil {
		return nil, err
	}

	if user.UserID == 0 {
		return nil, errorUnauthorize
	}

	hasher := sha1.New()
	hasher.Write([]byte(password))
	sum := hasher.Sum([]byte(srv.passwordSalt))

	if !bytes.Equal(sum, user.Password) {
		return nil, errorUnauthorize
	}

	tokens, tokenUUID, err := srv.tokenGenerate(user.UserID)
	if err != nil {
		return nil, err
	}

	_, err = pg.Exec("update profiles.users set token_uuid=$1,last_login_dt=now() where user_id=$2", tokenUUID, user.UserID)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func ValidateAccessToken(tokenStr string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		cfg := config.GetInstance().App

		return []byte(cfg.JwtAccessSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("incorrect token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("suspicious token")
	}

	tokenExp := int64(claims["exp"].(float64))

	nowTime := time.Now().Unix()
	if nowTime > tokenExp {
		return nil, errors.New("you're unauthorized")
	}

	return &claims, nil
}

func (srv *AuthService) tokenGenerate(userID int64) (*AuthResultModel, string, error) {
	tokenUUID := uuid.New().String()

	token := jwt.New(jwt.SigningMethodHS256)
	accessTokenExp := time.Now().Add(20 * time.Minute).Unix()
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = accessTokenExp
	claims["uid"] = userID
	claims["uuid"] = tokenUUID

	tokenString, err := token.SignedString([]byte(srv.jwtTokenSecret))
	if err != nil {
		return nil, "", err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenExp := time.Now().Add(24 * 30 * 6 * time.Hour).Unix()
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["exp"] = refreshTokenExp
	refreshClaims["uid"] = userID
	refreshClaims["uuid"] = tokenUUID

	refreshTokenString, err := token.SignedString([]byte(srv.jwtRefreshSecret))
	if err != nil {
		return nil, tokenUUID, err
	}

	result := AuthResultModel{
		AccessToken:     tokenString,
		AccessTokenExp:  accessTokenExp,
		RefreshToken:    refreshTokenString,
		RefreshTokenExp: refreshTokenExp,
	}
	return &result, tokenUUID, nil
}
