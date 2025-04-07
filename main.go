package main

import (
	"fmt"

	"github.com/ezydark/force/libs/logger"
	"github.com/ezydark/force/libs/util"
	"github.com/ezydark/force/libs/warp"
	"github.com/ezydark/force/libs/win"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	err := logger.Init()
	if err != nil {
		fatal_tag := color.New(color.FgRed, color.Bold).Sprintf("[FATAL]")
		fmt.Println(fatal_tag, "Could not initialize custom logger:", err)
		return
	}
	log.Info().Msg(color.New(color.Bold).Sprintf("WarpEnforcer starting..."))
	util.WaitForInput()

	// Ensure to run myself as admin
	err = win.Admin.EnsureSelfAdmin()
	if err != nil {
		log.Fatal().Msgf("Could not ensure if I ran as admin:\n %v", err)
	}

	// Check if Warp service is installed
	warpInstalled, err := warp.IsInstalled()
	if err != nil {
		log.Fatal().Msgf("Could not ensure Warp is installed:\n %v", err)
	}
	if !warpInstalled {
		log.Fatal().Msg("Warp is not properly installed! Install it using package manager like 'winget' or other.")
	} else {
		log.Info().Msg("Warp is installed")
	}

	// Check if Warp service is enabled for startup and running
	serv, err := warp.Serv.Init()
	if err != nil {
		log.Fatal().Msgf("Could not initialize Windows service manager with Warp service:\n %v", err)
	}
	defer serv.Close()
	err = serv.EnsureIsEnabled()
	if err != nil {
		log.Fatal().Msgf("Could not ensure Warp service is enabled for startup:\n %v", err)
	} else {
		log.Info().Msg("Warp service is enabled")
	}
	err = serv.EnsureIsRunning()
	if err != nil {
		log.Fatal().Msgf("Could not ensure Warp service is running:\n %v", err)
	} else {
		log.Info().Msg("Warp service is running")
	}

	// Check if Warp service is connected to the Cloudflare service
	warpConnected, err := warp.IsConnected()
	if err != nil {
		log.Fatal().Msgf("Could not check Warp connection state to Cloudflare service:\n %v", err)
	}
	if !warpConnected {
		log.Error().Msg("Warp is not connected to the Cloudflare service! Trying to connect again...")
		err = warp.Connect()
		if err != nil {
			log.Fatal().Msgf("Could not connect Warp to the Cloudflare service:\n %v", err)
		}
	}
	log.Info().Msg("Warp is connected to the Cloudflare service")

	// Prevent app from being closed at the end
	util.WaitForInput()
}
