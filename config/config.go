package config

import (
	"encoding/json"
	"os"
)

const configFilName string = "config.json"

type Oauth struct {
	TokenURL     string `json:"TokenURL"`
	GrantType    string `json:"GrantType"`
	ClientID     string `json:"ClientID"`
	ClientSecret string `json:"ClientSecret"`
	RedirectURI  string `json:"RedirectURI"`
}

// Configuration структура конфига
type Configuration struct {
	App struct {
		ServerAddr               string `json:"server_addr"`
		ImageMagick              string `json:"imageMagick"`
		JwtAccessPrivateKeyPath  string `json:"jwt_access_private_key_path"`
		JwtAccessPublicKeyPath   string `json:"jwt_access_public_key_path"`
		JwtRefreshPrivateKeyPath string `json:"jwt_refresh_private_key_path"`
		JwtRefreshPublicKeyPath  string `json:"jwt_refresh_public_key_path"`
	} `json:"app"`
	SMTP struct {
		Host     string `json:"host"`
		Port     uint16 `json:"port"`
		From     string `json:"from"`
		Password string `json:"password"`
	} `json:"smtp"`
	Memcache struct {
		Addr string `json:"addr"`
	} `json:"memcache"`
	Db struct {
		DriverName       string `json:"driver_name"`
		ConnectionString string `json:"connection_string"`
		MaxIdleConns     int    `json:"max_idle_conns"`
		MaxOpenConns     int    `json:"max_open_conns"`
	} `json:"db"`
	Oauth struct {
		MailRu Oauth `json:"mailru"`
		Yandex Oauth `json:"yandex"`
	} `json:"oauth"`
}

var cfg *Configuration

// ReadConfig чтение файла настроек
func ReadConfig() Configuration {
	cfg = new(Configuration)
	cfg.read()
	return *cfg
}

// GetInstance получить настройки
func GetInstance() Configuration {
	return *cfg
}

func (cfg *Configuration) read() {
	body, err := os.ReadFile(configFilName)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &cfg)
	if err != nil {
		panic(err)
	}
}
