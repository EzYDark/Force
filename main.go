package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ezydark/warpenforcer/libs/appconfig"
	"github.com/ezydark/warpenforcer/libs/logger"
	"github.com/ezydark/warpenforcer/libs/warpservice"
	"github.com/fatih/color"
)

func waitForInputToEnd() error {
	_, err := fmt.Print(color.New(color.FgRed).Sprint("\nEverything is done!\n--- Press Enter to close this window..."))
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

	// Ensure if running with admin rights
	// if err = admin.EnsureAdmin(); err != nil {
	// 	log.Fatal().Msgf("Could not ensure if I ran as admin:\n %v", err)
	// }

	// Prepare configuration
	_, err = appconfig.Init()
	if err != nil {
		log.Fatal().Msgf("Could not initialize appconfig:\n %v", err)
	}

	// Check for existing and running Warp
	warp, err := warpservice.Init()
	if err != nil {
		log.Fatal().Msgf("Could not initialize WarpService:\n %v", err)
	}
	err = warp.EnsureIsRunning()
	if err != nil {
		log.Fatal().Msgf("Could not ensure Warp service is running:\n %v", err)
	}

	// Prevent app from being closed at the end
	waitForInputToEnd()
}
