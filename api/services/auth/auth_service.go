package auth_service

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"odo24_mobile_backend/api/services"
	"odo24_mobile_backend/api/utils"
	"odo24_mobile_backend/db"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	defaultAccessTokenExp  = time.Minute * 20
	defaultRefreshTokenExp = time.Hour * 24 * 30 * 6
)

type AuthService struct {
	jwtAccessPrivateKey  []byte
	jwtAccessPublicKey   []byte
	jwtRefreshPrivateKey []byte
	jwtRefreshPublicKey  []byte
}

func NewAuthService(jwtAccessPrivateKeyPath, jwtAccessPublicKeyPath, jwtRefreshPrivateKeyPath, jwtRefreshPublicKeyPath string) *AuthService {
	// jwt access
	accessPrivateKey, err := os.ReadFile(jwtAccessPrivateKeyPath)
	if err != nil {
		panic(err)
	}
	accessPublicKey, err := os.ReadFile(jwtAccessPublicKeyPath)
	if err != nil {
		panic(err)
	}
	// jwt refresh
	refreshPrivateKey, err := os.ReadFile(jwtRefreshPrivateKeyPath)
	if err != nil {
		panic(err)
	}
	refreshPublicKey, err := os.ReadFile(jwtRefreshPublicKeyPath)
	if err != nil {
		panic(err)
	}

	return &AuthService{
		jwtAccessPrivateKey:  accessPrivateKey,
		jwtAccessPublicKey:   accessPublicKey,
		jwtRefreshPrivateKey: refreshPrivateKey,
		jwtRefreshPublicKey:  refreshPublicKey,
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

	if !bytes.Equal(currentPassword, user.Password) {
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

	newSalt, err := utils.GenerateSalt()
	if err != nil {
		return err
	}

	_, err = pg.Exec("update profiles.users set password_hash=$1,salt=$2 where user_id=$3", newHashPassword, newSalt, userID)
	if err != nil {
		return err
	}

	return nil
}

/*
RefreshToken рефреш токена
*/
func (srv *AuthService) RefreshToken(accessTokenStr, refreshTokenStr string) (*AuthResultModel, error) {
	accessToken, err := srv.ParseAccessToken(accessTokenStr, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, err
	}

	refreshToken, err := srv.ParseRefreshToken(refreshTokenStr)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, services.ErrorUnauthorize
		}
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
		log.Printf("accessUUID=%s, refreshUUID=%s", accessUUID, refreshUUID)
		return nil, errors.New("accessUUID not equal refreshUUID")
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
	return tokens, err
}

func (srv *AuthService) ParseAccessToken(tokenString string, options ...jwt.ParserOption) (*jwt.Token, error) {
	return srv.parseToken(tokenString, srv.jwtAccessPublicKey, options...)
}

func (srv *AuthService) ParseRefreshToken(tokenString string, options ...jwt.ParserOption) (*jwt.Token, error) {
	return srv.parseToken(tokenString, srv.jwtRefreshPublicKey, options...)
}

func (srv *AuthService) parseToken(tokenString string, pubKey []byte, options ...jwt.ParserOption) (*jwt.Token, error) {
	key, err := jwt.ParseRSAPublicKeyFromPEM(pubKey)
	if err != nil {
		log.Printf("error parsing RSA public key: %v\n", err)
		return nil, services.ErrorParsingRSAPublicKey
	}

	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	}, options...)
	return parsedToken, err
}

func (srv *AuthService) tokenGenerate(userID int64) (*AuthResultModel, string, error) {
	tokenUUID := uuid.New().String()

	// access
	accessKey, err := jwt.ParseRSAPrivateKeyFromPEM(srv.jwtAccessPrivateKey)
	if err != nil {
		log.Printf("error parsing RSA private key: %v\n", err)
		return nil, "", services.ErrorParsingRSAPrivateKey
	}

	accessToken := jwt.New(jwt.SigningMethodRS256)

	accessTokenExp := time.Now().Add(defaultAccessTokenExp).Unix()
	accessClaims := accessToken.Claims.(jwt.MapClaims)
	accessClaims["exp"] = accessTokenExp
	accessClaims["uid"] = userID
	accessClaims["uuid"] = tokenUUID

	accessTokenString, err := accessToken.SignedString(accessKey)
	if err != nil {
		log.Printf("error signing token: %v\n", err)
		return nil, tokenUUID, services.ErrorSigningJwtToken
	}

	// refresh
	refreshKey, err := jwt.ParseRSAPrivateKeyFromPEM(srv.jwtRefreshPrivateKey)
	if err != nil {
		log.Printf("error parsing RSA private key: %v\n", err)
		return nil, "", services.ErrorParsingRSAPrivateKey
	}

	refreshToken := jwt.New(jwt.SigningMethodRS256)

	refreshTokenExp := time.Now().Add(defaultRefreshTokenExp).Unix()
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["exp"] = refreshTokenExp
	refreshClaims["uuid"] = tokenUUID

	refreshTokenString, err := refreshToken.SignedString(refreshKey)
	if err != nil {
		log.Printf("error signing token: %v\n", err)
		return nil, tokenUUID, services.ErrorSigningJwtToken
	}

	result := AuthResultModel{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}
	return &result, tokenUUID, nil
}
