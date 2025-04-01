package warpservice

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ezydark/warpenforcer/libs/appconfig"
	"github.com/ezydark/warpenforcer/libs/fs"
	"github.com/ezydark/warpenforcer/libs/logger"
	"github.com/ezydark/warpenforcer/libs/processutil"
	"github.com/fatih/color"
	"github.com/rs/zerolog"
)

type WarpService struct {
	log    *zerolog.Logger
	config *appconfig.AppConfig
}

var warp *WarpService

// Initialize WarpService
func Init() (*WarpService, error) {
	if warp != nil {
		return nil, errors.New("warp is already initialized")
	}

	log, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("Not able to get Logger: %w", err)
	}
	config, err := appconfig.Get()
	if err != nil {
		return nil, fmt.Errorf("Not able to get AppConfig: %w", err)
	}

	warp = &WarpService{
		log:    log,
		config: config,
	}

	return warp, nil
}

// Get WarpService pointer
func Get() (*WarpService, error) {
	if warp == nil {
		return nil, errors.New("warp is not initialized")
	}
	return warp, nil
}

func (warpserv *WarpService) EnsureIsRunning() error {
	guiProcessPath := warpserv.config.WarpFolderPath + "\\" + warpserv.config.WarpProcessName

	// Check if Warp is installed
	warpInstalled, err := warpserv.IsInstalled()
	if err != nil {
		return fmt.Errorf("Could not check if Warp is installed:\n %w", err)
	}
	if !warpInstalled {
		return errors.New(color.New(color.FgGreen, color.Bold).Sprintf("Warp's folder cannot be found.\n Clouflare Warp is probably not installed.\n Install it using 'winget install Cloudflare.Warp'"))
	}

	// Check if Warp process is running
	warpRunning, err := warpserv.IsRunning()
	if err != nil {
		return fmt.Errorf("error while checking if Warp is running:\n %w", err)
	}
	if warpRunning {
		warpserv.log.Info().Msgf("'%v' is running", warpserv.config.WarpProcessName)
	} else {
		warpserv.log.Error().Msgf("'%v' is NOT running! Trying to start it again...", warpserv.config.WarpProcessName)
		err := processutil.StartProcess(guiProcessPath)
		if err != nil {
			return fmt.Errorf("error while starting Warp:\n %w", err)
		}
	}

	return warpserv.WaitForProcessToStart(20, 500*time.Millisecond)
}

// Check if Warp is installed
func (warpserv *WarpService) IsInstalled() (bool, error) {
	// Check if Warp's folder exists
	warpDirExists, err := fs.DirExists(warpserv.config.WarpFolderPath)
	if err != nil {
		return false, fmt.Errorf("Could not check if Warp's folder exists:\n %w", err)
	}

	// Check if Warp GUI executable exists
	warpGuiExists, err := fs.FileExists(warpserv.config.WarpFolderPath + "\\" + warpserv.config.WarpProcessName)
	if err != nil {
		return false, fmt.Errorf("Could not check if Warp GUI exists:\n %w", err)
	}

	if !warpDirExists || !warpGuiExists {
		return false, nil
	} else {
		return true, nil
	}
}

// Check if Warp process is running
func (warpserv *WarpService) IsRunning() (bool, error) {
	warpRunning, err := processutil.IsProcessRunningByName(warpserv.config.WarpProcessName)
	if err != nil {
		return false, fmt.Errorf("error searching for '%v' process:\n %w", warpserv.config.WarpProcessName, err)
	}
	if warpRunning {
		return true, nil
	} else {
		return false, nil
	}
}

// Check if Warp process is connected
func (warpserv *WarpService) IsConnected() (bool, error) {
	cmd := exec.Command("warp-cli", "status")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("error checking Warp status:\n %w", err)
	}

	if strings.Contains(string(out), "Connected") {
		return true, nil
	} else {
		return false, nil
	}
}

// Connect Warp to the Cloudflare service
func (warpserv *WarpService) Connect() (bool, error) {
	cmd := exec.Command("warp-cli", "connect")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("error connecting Warp to Cloudflare service:\n %w", err)
	}

	if strings.Contains(string(out), "Success") {
		return true, nil
	} else {
		return false, nil
	}
}

// Wait for Warp process to start
func (warpserv *WarpService) WaitForProcessToStart(maxAttempts int, waitTime time.Duration) error {
	warpRunning, err := warpserv.IsRunning()
	if err != nil {
		return fmt.Errorf("could not check if '%v' is running:\n %w", warpserv.config.WarpProcessName, err)
	}

	if warpRunning {
		return nil
	}

	for attempt := 1; attempt < maxAttempts; attempt++ {
		warpserv.log.Debug().Msgf("[%v/%v] Waiting for '%v' to start...",
			attempt, maxAttempts, warpserv.config.WarpProcessName)
		time.Sleep(waitTime)

		warpRunning, err = warpserv.IsRunning()
		if err != nil {
			return fmt.Errorf("could not check if '%v' is running:\n %w",
				warpserv.config.WarpProcessName, err)
		}

		if warpRunning {
			warpserv.log.Info().Msgf("'%v' successfully started after %v attempts",
				warpserv.config.WarpProcessName, attempt)
			return nil
		}
	}

	return fmt.Errorf("could not start '%v' after %v attempts",
		warpserv.config.WarpProcessName, maxAttempts)
}
