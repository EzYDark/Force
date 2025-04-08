package warp

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ezydark/ezforce/app/config"
	"github.com/ezydark/ezforce/libs/warp/serv"
	"github.com/ezydark/ezforce/libs/win"
	"github.com/rs/zerolog/log"
)

var Serv *serv.WarpServ

// Check if Warp executables are installed
func IsInstalled() (bool, error) {
	// Check if Warp's folder exists
	warpDirExists, err := win.Fs.DirExists(config.Warp.FolderPath)
	if err != nil {
		return false, fmt.Errorf("Could not check if Warp's folder exists:\n %w", err)
	}

	// Check if Warp GUI executable exists
	warpGuiExists, err := win.Fs.FileExists(config.Warp.FolderPath + "\\" + config.Warp.GUIExecName)
	if err != nil {
		return false, fmt.Errorf("Could not check if Warp GUI exists:\n %w", err)
	}

	// Check if 'warp-svc.exe' exists
	warpSvcExists, err := win.Fs.FileExists(config.Warp.FolderPath + "\\" + config.Warp.SvcExecName)
	if err != nil {
		return false, fmt.Errorf("Could not check if Warp Svc exists:\n %w", err)
	}

	if !warpDirExists || !warpGuiExists || !warpSvcExists {
		return false, nil
	} else {
		return true, nil
	}
}

// Check if Warp process is connected to the Cloudflare service
func IsConnected() (bool, error) {
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
func Connect() error {
	cmd := exec.Command("warp-cli", "connect")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error connecting Warp to Cloudflare service:\n %w", err)
	}

	if strings.Contains(string(out), "Success") {
		return nil
	} else {
		return waitForWarpToConnect(20, 500*time.Millisecond)
	}
}

func EnsureIsConnected() error {
	isConnected, err := IsConnected()
	if err != nil {
		return fmt.Errorf("could not check Warp connection state to Cloudflare service:\n %v", err)
	}
	if !isConnected {
		err = Connect()
		if err != nil {
			return fmt.Errorf("could not connect Warp to the Cloudflare service:\n %v", err)
		}
	}
	return nil
}

func waitForWarpToConnect(maxAttempts int, waitTime time.Duration) error {
	connected, err := IsConnected()
	if err != nil {
		return fmt.Errorf("could not check Warp connection state to Cloudflare service:\n %v", err)
	}
	if connected {
		return nil
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Debug().Msgf("[%v/%v] Waiting for Warp to connect to Cloudflare service...", attempt, maxAttempts)
		time.Sleep(waitTime)

		connected, err = IsConnected()
		if err != nil {
			return fmt.Errorf("could not check Warp connection state to Cloudflare service:\n %v", err)
		}
		if connected {
			log.Debug().Msgf("Warp connected to Cloudflare service after %v attempts", attempt)
			return nil
		}
	}

	return fmt.Errorf("could not connect Warp to the Cloudflare service after %d attempts", maxAttempts)
}
