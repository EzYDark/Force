package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ezydark/warpenforcer/libs/appconfig"
	"github.com/ezydark/warpenforcer/libs/logger"
	warp_exec "github.com/ezydark/warpenforcer/libs/warp/exec"
	warp_serv "github.com/ezydark/warpenforcer/libs/warp/serv"
	"github.com/fatih/color"
)

func waitForInput() error {
	_, err := fmt.Print(color.New(color.FgRed).Sprint("-- Press Enter to continue..."))
	if err != nil {
		return fmt.Errorf("Could not Print:\n %w", err)
	}

	reader := bufio.NewReader(os.Stdin)
	_, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("Could not ReadString:\n %w", err)
	}

	return nil
}

func main() {
	// Initialize logger
	log, err := logger.Init()
	if err != nil {
		fatal_tag := color.New(color.FgRed).Sprintf("[FATAL]")
		fmt.Println(fatal_tag, "Could not initialize logger:", err)
		return
	}
	log.Info().Msg(color.New(color.Bold).Sprintf("WarpEnforcer starting..."))
	waitForInput()

	// Ensure to run myself as admin
	// if err = admin.EnsureAdmin(); err != nil {
	// 	log.Fatal().Msgf("Could not ensure if I ran as admin:\n %v", err)
	// }

	// Prepare configuration
	_, err = appconfig.Init()
	if err != nil {
		log.Fatal().Msgf("Could not initialize appconfig:\n %v", err)
	}

	// Check for existing and running Warp executable
	warpexec, err := warp_exec.Init()
	if err != nil {
		log.Fatal().Msgf("Could not initialize WarpService:\n %v", err)
	}
	err = warpexec.EnsureIsRunning()
	if err != nil {
		log.Fatal().Msgf("Could not ensure Warp service is running:\n %v", err)
	}

	// Initialize Warp service manager, open the specific service, and get the service's configuration
	warpserv, err := warp_serv.Init()
	if err != nil {
		log.Fatal().Msgf("Could not initialize Warp service manager:\n %v", err)
	}
	defer warpserv.Close()

	// Check if Warp service is enabled for startup
	err = warpserv.EnsureIsEnabled()
	if err != nil {
		log.Fatal().Msgf("Could not ensure Warp service is enabled:\n %v", err)
	} else {
		log.Info().Msg("Warp service is enabled for startup")
	}

	// Prevent app from being closed at the end
	waitForInput()
}
