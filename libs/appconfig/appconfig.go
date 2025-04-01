package appconfig

import (
	"errors"
)

type AppConfig struct {
	WarpFolderPath  string
	WarpProcessName string
}

var config *AppConfig

func Init() (*AppConfig, error) {
	if config != nil {
		return nil, errors.New("config is already initialized")
	}

	config = &AppConfig{
		WarpFolderPath:  "C:\\Program Files\\Cloudflare\\Cloudflare WARP",
		WarpProcessName: "Cloudflare WARP.exe",
	}
	return config, nil
}

func Get() (*AppConfig, error) {
	if config == nil {
		return nil, errors.New("config is not initialized")
	}
	return config, nil
}
