package exec

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

type WarpExec struct {
	log     *zerolog.Logger
	appconf *appconfig.AppConfig
}

var warp_exec *WarpExec

// Initialize WarpExec
func Init() (*WarpExec, error) {
	if warp_exec != nil {
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

	warp_exec = &WarpExec{
		log:     log,
		appconf: config,
	}

	return warp_exec, nil
}

// Get WarpExec pointer
func Get() (*WarpExec, error) {
	if warp_exec == nil {
		return nil, errors.New("WarpExec is not initialized")
	}
	return warp_exec, nil
}

// Ensure that Warp executable is running
func (warpexec *WarpExec) EnsureIsRunning() error {
	guiProcessPath := warpexec.appconf.WarpFolderPath + "\\" + warpexec.appconf.WarpProcessName

	// Check if Warp is installed
	warpInstalled, err := warpexec.IsInstalled()
	if err != nil {
		return fmt.Errorf("Could not check if Warp is installed:\n %w", err)
	}
	if !warpInstalled {
		return errors.New(color.New(color.FgGreen, color.Bold).Sprintf("Warp's folder cannot be found.\n Clouflare Warp is probably not installed.\n Install it using 'winget install Cloudflare.Warp'"))
	}

	// Check if Warp process is running
	warpRunning, err := warpexec.IsRunning()
	if err != nil {
		return fmt.Errorf("error while checking if Warp is running:\n %w", err)
	}
	if warpRunning {
		warpexec.log.Info().Msgf("'%v' is running", warpexec.appconf.WarpProcessName)
		return nil
	} else {
		warpexec.log.Error().Msgf("'%v' is NOT running! Trying to start it again...", warpexec.appconf.WarpProcessName)
		err := processutil.StartProcess(guiProcessPath)
		if err != nil {
			return fmt.Errorf("error while starting Warp:\n %w", err)
		}
		return warpexec.waitForWarpToStart(20, 500*time.Millisecond)
	}
}

// Check if Warp executable is installed
func (warpexec *WarpExec) IsInstalled() (bool, error) {
	// Check if Warp's folder exists
	warpDirExists, err := fs.DirExists(warpexec.appconf.WarpFolderPath)
	if err != nil {
		return false, fmt.Errorf("Could not check if Warp's folder exists:\n %w", err)
	}

	// Check if Warp GUI executable exists
	warpGuiExists, err := fs.FileExists(warpexec.appconf.WarpFolderPath + "\\" + warpexec.appconf.WarpProcessName)
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
func (warpexec *WarpExec) IsRunning() (bool, error) {
	warpRunning, err := processutil.IsProcessRunningByName(warpexec.appconf.WarpProcessName)
	if err != nil {
		return false, fmt.Errorf("error searching for '%v' process:\n %w", warpexec.appconf.WarpProcessName, err)
	}
	if warpRunning {
		return true, nil
	} else {
		return false, nil
	}
}

// Check if Warp process is connected
func (warpexec *WarpExec) IsConnected() (bool, error) {
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
func (warpexec *WarpExec) Connect() (bool, error) {
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
func (warpexec *WarpExec) waitForWarpToStart(maxAttempts int, waitTime time.Duration) error {
	warpRunning, err := warpexec.IsRunning()
	if err != nil {
		return fmt.Errorf("could not check if '%v' is running:\n %w", warpexec.appconf.WarpProcessName, err)
	}

	if warpRunning {
		return nil
	}

	for attempt := 1; attempt < maxAttempts; attempt++ {
		warpexec.log.Debug().Msgf("[%v/%v] Waiting for '%v' to start...",
			attempt, maxAttempts, warpexec.appconf.WarpProcessName)
		time.Sleep(waitTime)

		warpRunning, err = warpexec.IsRunning()
		if err != nil {
			return fmt.Errorf("could not check if '%v' is running:\n %w",
				warpexec.appconf.WarpProcessName, err)
		}

		if warpRunning {
			warpexec.log.Info().Msgf("'%v' successfully started after %v attempts",
				warpexec.appconf.WarpProcessName, attempt)
			return nil
		}
	}

	return fmt.Errorf("could not start '%v' after %v attempts",
		warpexec.appconf.WarpProcessName, maxAttempts)
}
