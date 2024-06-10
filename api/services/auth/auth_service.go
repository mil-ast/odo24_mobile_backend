package auth_service

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"odo24_mobile_backend/api/services"
	"odo24_mobile_backend/api/utils"
	"odo24_mobile_backend/config"
	"odo24_mobile_backend/db"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	defaultAccessTokenExp  = time.Minute * 5 // TODO
	defaultRefreshTokenExp = 24 * 30 * 6 * time.Hour
)

type AuthService struct {
	jwtTokenSecret   string
	jwtRefreshSecret string
	passwordSalt     []byte
}

func NewAuthService(jwtTokenSecret, jwtRefreshSecret, passwordSalt string) *AuthService {
	return &AuthService{
		jwtTokenSecret:   jwtTokenSecret,
		jwtRefreshSecret: jwtRefreshSecret,
		passwordSalt:     []byte(passwordSalt),
	}
}

func (srv *AuthService) Login(email string, password string) (*AuthResultModel, error) {
	pg := db.Conn()
	var user struct {
		UserID   int64
		Password []byte
		Salt     []byte
	}
	err := pg.QueryRow("select u.user_id,u.password_hash,u.salt from profiles.users u where u.login=$1", email).Scan(&user.UserID, &user.Password, &user.Salt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrorUnauthorize
		}
		return nil, err
	}

	if user.UserID == 0 {
		return nil, services.ErrorUnauthorize
	}

	currentPassword, err := utils.GetPasswordHash([]byte(password), user.Salt)
	if err != nil {
		return nil, err
	}

	userPassword, err := utils.GetPasswordHash(user.Password, user.Salt)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(currentPassword, userPassword) {
		return nil, services.ErrorUnauthorize
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
	var currentSalt []byte
	err := pg.QueryRow("select u.password_hash,u.salt from profiles.users u where u.user_id=$1", userID).Scan(&currentPassword, &currentSalt)
	if err != nil {
		return err
	}

	oldHashPassword, err := utils.GetPasswordHash([]byte(oldPassword), currentSalt)
	if err != nil {
		return err
	}

	currentHashPassword, err := utils.GetPasswordHash(currentPassword, currentSalt)
	if err != nil {
		return err
	}

	if !bytes.Equal(oldHashPassword, currentHashPassword) {
		return errors.New("invalid password")
	}

	salt, err := utils.GenerateSalt()
	if err != nil {
		return err
	}

	newHashPassword, err := utils.GetPasswordHash([]byte(newPassword), salt)
	if err != nil {
		return err
	}

	_, err = pg.Exec("update profiles.users set password_hash=$1 where user_id=$2", newHashPassword, userID)
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
		return nil, err
	}

	refreshToken, err := getToken(refreshTokenStr, []byte(srv.jwtRefreshSecret), jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, err
	}

	accessClaims, ok := accessToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("AccessTokenClaimsIsEmpty")
	}

	refreshClaims, ok := refreshToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("RefreshTokenClaimsIsEmpty")
	}

	accessUUID := accessClaims["uuid"].(string)
	refreshUUID := refreshClaims["uuid"].(string)
	if accessUUID != refreshUUID {
		return nil, errors.New("accessUUID not equal refreshUUID")
	}

	// проверка, что рефреш токен не протух
	refreshTokenExp := int64(refreshClaims["exp"].(float64))
	nowTime := time.Now().Unix()
	if nowTime > refreshTokenExp {
		return nil, services.ErrorUnauthorize
	}

	userID := int64(accessClaims["uid"].(float64))

	pg := db.Conn()
	var dbRefreshUUID string
	err = pg.QueryRow("select u.token_uuid from profiles.users u where u.user_id=$1", userID).Scan(&dbRefreshUUID)
	if err != nil {
		return nil, err
	}

	if dbRefreshUUID != refreshUUID {
		return nil, services.ErrorUnauthorize
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
	accessTokenExp := time.Now().Add(defaultAccessTokenExp).Unix()
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = accessTokenExp
	claims["uid"] = userID
	claims["uuid"] = tokenUUID

	tokenString, err := token.SignedString([]byte(srv.jwtTokenSecret))
	if err != nil {
		return nil, "", err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenExp := time.Now().Add(defaultRefreshTokenExp).Unix()
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["exp"] = refreshTokenExp
	refreshClaims["uid"] = userID
	refreshClaims["uuid"] = tokenUUID

	refreshTokenString, err := refreshToken.SignedString([]byte(srv.jwtRefreshSecret))
	if err != nil {
		return nil, tokenUUID, err
	}

	result := AuthResultModel{
		AccessToken:  tokenString,
		RefreshToken: refreshTokenString,
	}
	return &result, tokenUUID, nil
}
