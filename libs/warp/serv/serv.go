package serv

import (
	"errors"
	"fmt"
	"time"

	"github.com/ezydark/ezforce/app/config"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type WarpServ struct {
	ServMgr  *mgr.Mgr
	WarpServ *mgr.Service
}

// Initialize the Windows service manager with Warp service
func (s *WarpServ) Init() (*WarpServ, error) {
	if s != nil {
		return nil, errors.New("warpserv is already initialized")
	}

	// Connect to the Windows service manager
	manager, err := mgr.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to service manager:\n %w", err)
	}

	// Open the specific service
	warp_service_name := config.Warp.ServiceName
	service, err := manager.OpenService(warp_service_name)
	if err != nil {
		return nil, fmt.Errorf("failed to open service '%s':\n %w", warp_service_name, err)
	}

	s = &WarpServ{
		ServMgr:  manager,
		WarpServ: service,
	}

	return s, nil
}

// Close the Warp service and Windows service manager
func (s *WarpServ) Close() error {
	if s != nil {
		s.WarpServ.Close()
		s.WarpServ = nil
		s.ServMgr.Disconnect()
		s.ServMgr = nil
	}
	return nil
}

// Ensure that the Warp service is set to startup automatically
func (s *WarpServ) EnsureIsEnabled() error {
	enabled, err := s.IsEnabled()
	if err != nil {
		return err
	}
	if enabled {
		return nil
	} else {
		log.Error().Msg("Warp service is not enabled for startup! Trying to enable it...")
		return s.Enable()
	}
}

func (s *WarpServ) EnsureIsRunning() error {
	isRunning, err := s.IsRunning()
	if err != nil {
		return err
	}
	if isRunning {
		return nil
	} else {
		log.Error().Msg("Warp service is not running! Trying to start it...")
		return s.Start()
	}
}

// Check if the Warp service is set to startup automatically
func (s *WarpServ) IsEnabled() (bool, error) {
	serv_conf, err := s.WarpServ.Config()
	if err != nil {
		return false, fmt.Errorf("failed to get service config:\n %w", err)
	}

	if serv_conf.StartType == mgr.StartAutomatic {
		return true, nil
	} else {
		return false, nil
	}
}

func (s *WarpServ) Enable() error {
	serv_conf, err := s.WarpServ.Config()
	if err != nil {
		return fmt.Errorf("failed to get service config:\n %w", err)
	}

	if serv_conf.StartType == mgr.StartAutomatic {
		return nil
	}

	newConfig := mgr.Config{
		StartType: mgr.StartAutomatic,
		// Keep other settings the same
		DisplayName:      serv_conf.DisplayName,
		Description:      serv_conf.Description,
		BinaryPathName:   serv_conf.BinaryPathName,
		LoadOrderGroup:   serv_conf.LoadOrderGroup,
		Dependencies:     serv_conf.Dependencies,
		ServiceStartName: serv_conf.ServiceStartName,
		DelayedAutoStart: serv_conf.DelayedAutoStart,
		ErrorControl:     serv_conf.ErrorControl,
		ServiceType:      serv_conf.ServiceType,
	}

	err = s.WarpServ.UpdateConfig(newConfig)
	if err != nil {
		return fmt.Errorf("failed to update service config: %w", err)
	}

	return s.waitForWarpServToBeEnabled(20, 500*time.Millisecond)
}

// Check if the Warp service's status is running
func (s *WarpServ) IsRunning() (bool, error) {
	status, err := s.WarpServ.Query()
	if err != nil {
		return false, fmt.Errorf("failed to get service status:\n %w", err)
	}

	isRunning := status.State == svc.Running

	return isRunning, nil
}

func (s *WarpServ) Start() error {
	err := s.WarpServ.Start()
	if err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return s.waitForWarpServToBeRunning(20, 500*time.Millisecond)
}

func (s *WarpServ) waitForWarpServToBeRunning(maxAttempts int, waitTime time.Duration) error {
	isRunning, err := s.IsRunning()
	if err != nil {
		return fmt.Errorf("failed to check Warp service status: %w", err)
	}
	if isRunning {
		return nil
	}

	for attempt := 1; attempt < maxAttempts; attempt++ {
		log.Debug().Msgf("[%v/%v] Waiting for '%v' to start...",
			attempt, maxAttempts, config.Warp.ServiceName)
		isRunning, err := s.IsRunning()
		if err != nil {
			return fmt.Errorf("failed to check Warp service status: %w", err)
		}
		if isRunning {
			log.Debug().Msgf("'%v' started successfully", config.Warp.ServiceName)
			return nil
		}
		time.Sleep(waitTime)
	}
	return fmt.Errorf("failed to start service after %v attempts", maxAttempts)
}

// Wait for Warp service to start
func (s *WarpServ) waitForWarpServToBeEnabled(maxAttempts int, waitTime time.Duration) error {
	enabled, err := s.IsEnabled()
	if err != nil {
		return fmt.Errorf("failed to check if Warp service is enabled for startup: %w", err)
	}
	if enabled {
		return nil
	}

	for attempt := 1; attempt < maxAttempts; attempt++ {
		log.Debug().Msgf("[%v/%v] Waiting for '%v' to start...",
			attempt, maxAttempts, config.Warp.ServiceName)
		time.Sleep(waitTime)

		enabled, err = s.IsEnabled()
		if err != nil {
			return fmt.Errorf("failed to check if Warp service is enabled for startup: %w", err)
		}
		if enabled {
			log.Debug().Msgf("'%v' successfully started after %v attempts",
				config.Warp.ServiceName, attempt)
			return nil
		}
	}

	return fmt.Errorf("could not enable '%v' for startup after %v attempts",
		config.Warp.ServiceName, maxAttempts)
}
