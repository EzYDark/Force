package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type AppConfig struct {
	InstallPath string `json:"installPath"`
	ExecName    string `json:"execName"`
	LogFileName string `json:"logFileName"`
	ConfigName  string `json:"configName"`
	ServiceName string `json:"serviceName"`
}

type WarpConfig struct {
	FolderPath  string `json:"folderPath"`
	GUIExecName string `json:"guiExecName"`
	SvcExecName string `json:"svcExecName"`
	ServiceName string `json:"serviceName"`
}

var App *AppConfig
var Warp *WarpConfig

type combinedConfigs struct {
	app  *AppConfig
	warp *WarpConfig
}

var configs *combinedConfigs

// Initialize the default configs
func init() {
	if App != nil {
		return
	}
	if Warp != nil {
		return
	}

	warp := &WarpConfig{}
	app := &AppConfig{}

	app.InstallPath = "C:\\Program Files\\ezForce"
	app.ExecName = "ezforce.exe"
	app.LogFileName = "ezforce.log"
	app.ConfigName = "ezforce.json"
	app.ServiceName = "ezForce"

	warp.FolderPath = "C:\\Program Files\\Cloudflare\\Cloudflare WARP"
	warp.GUIExecName = "Cloudflare WARP.exe"
	warp.SvcExecName = "warp-svc.exe"
	warp.ServiceName = "CloudflareWARP"

	App = app
	Warp = warp

	configs = &combinedConfigs{
		app:  App,
		warp: Warp,
	}

	return
}

// Load external config file
func Load(configPath string) error {
	// Check if default configs are initialized
	if App == nil {
		return errors.New("app config not initialized")
	}
	if Warp == nil {
		return errors.New("warp config not initialized")
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&configs); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}
