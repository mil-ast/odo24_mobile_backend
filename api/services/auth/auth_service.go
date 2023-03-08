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

var ErrorUnauthorize = errors.New("unauthorize")

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
		return nil, ErrorUnauthorize
	}

	hasher := sha1.New()
	hasher.Write([]byte(password))
	sum := hasher.Sum([]byte(srv.passwordSalt))

	if !bytes.Equal(sum, user.Password) {
		return nil, ErrorUnauthorize
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

func (srv *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	pg := db.Conn()
	var currentPassword []byte
	err := pg.QueryRow("select u.password_hash from profiles.users u where u.user_id = $1", userID).Scan(&currentPassword)
	if err != nil {
		return err
	}

	hasherOldPassword := sha1.New()
	_, err = hasherOldPassword.Write([]byte(oldPassword))
	if err != nil {
		return err
	}
	sumOldPasswd := hasherOldPassword.Sum([]byte(srv.passwordSalt))

	if !bytes.Equal(sumOldPasswd, currentPassword) {
		return errors.New("invalid password")
	}

	hasherNewPassword := sha1.New()
	hasherNewPassword.Write([]byte(newPassword))
	if err != nil {
		return err
	}
	sumNewPasswd := hasherNewPassword.Sum([]byte(srv.passwordSalt))

	_, err = pg.Exec("update profiles.users set password_hash=$1 where user_id=$2", sumNewPasswd, userID)
	if err != nil {
		return err
	}

	return nil
}

/*
RefreshToken рефреш токена
*/
func (srv *AuthService) RefreshToken(accessTokenStr, refreshTokenStr string) (*AuthResultModel, error) {
	accessToken, err := getToken(accessTokenStr, []byte(srv.jwtTokenSecret), jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, ErrorUnauthorize
	}

	refreshToken, err := getToken(refreshTokenStr, []byte(srv.jwtRefreshSecret), jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, ErrorUnauthorize
	}

	accessClaims, ok := accessToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrorUnauthorize
	}

	refreshClaims, ok := refreshToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrorUnauthorize
	}

	accessUUID := accessClaims["uuid"].(string)
	refreshUUID := refreshClaims["uuid"].(string)
	if accessUUID != refreshUUID {
		return nil, ErrorUnauthorize
	}

	// проверка, что рефреш токен не протух
	refreshTokenExp := int64(refreshClaims["exp"].(float64))
	nowTime := time.Now().Unix()
	if nowTime > refreshTokenExp {
		return nil, ErrorUnauthorize
	}

	userID := int64(accessClaims["uid"].(float64))

	pg := db.Conn()
	var dbRefreshUUID string
	err = pg.QueryRow("select u.token_uuid from profiles.users u where u.user_id=$1", userID).Scan(&dbRefreshUUID)
	if err != nil {
		return nil, err
	}

	if dbRefreshUUID != refreshUUID {
		return nil, ErrorUnauthorize
	}

	tokens, tokenUUID, err := srv.tokenGenerate(userID)
	if err != nil {
		return nil, err
	}

	_, err = pg.Exec("update profiles.users set token_uuid=$1 where user_id=$2", tokenUUID, userID)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func ValidateAccessToken(tokenStr string) (*jwt.MapClaims, error) {
	cfg := config.GetInstance().App
	token, err := getToken(tokenStr, []byte(cfg.JwtAccessSecret))
	if err != nil {
		return nil, err
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

func getToken(tokenStr string, secret []byte, options ...jwt.ParserOption) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	}, options...)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("incorrect token")
	}
	return token, nil
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

	refreshTokenString, err := refreshToken.SignedString([]byte(srv.jwtRefreshSecret))
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
