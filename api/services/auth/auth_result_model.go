package auth_service

type AuthResultModel struct {
	AccessToken     string `json:"access_token"`
	AccessTokenExp  int64  `json:"access_token_exp"`
	RefreshToken    string `json:"refresh_token"`
	RefreshTokenExp int64  `json:"refresh_token_exp"`
}
