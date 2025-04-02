package serv

import (
	"errors"
	"fmt"
	"time"

	"github.com/ezydark/warpenforcer/libs/appconfig"
	"github.com/ezydark/warpenforcer/libs/logger"
	"golang.org/x/sys/windows/svc/mgr"
)

type WarpServ struct {
	appconf     *appconfig.AppConfig
	mgr_service *mgr.Service
	mgr_manager *mgr.Mgr
}

var warp_serv *WarpServ

// Initialize the Warp service manager, open the specific service, and get the service's configuration
func Init() (*WarpServ, error) {
	if warp_serv != nil {
		return nil, errors.New("WarpServ is already initialized")
	}

	appconf, err := appconfig.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to load app configuration:\n %w", err)
	}

	// Connect to the Windows service manager
	manager, err := mgr.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to service manager:\n %w", err)
	}

	// Open the specific service
	service, err := manager.OpenService(appconf.WarpServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open service '%s':\n %w", appconf.WarpServiceName, err)
	}

	warp_serv = &WarpServ{
		appconf:     appconf,
		mgr_service: service,
		mgr_manager: manager,
	}

	return warp_serv, nil
}

func Get() (*WarpServ, error) {
	if warp_serv == nil {
		return nil, errors.New("WarpServ is not initialized")
	}
	return warp_serv, nil
}

// Close the Warp service manager and service
func (warpserv *WarpServ) Close() error {
	if warpserv.mgr_manager != nil {
		warpserv.mgr_manager.Disconnect()
		warpserv.mgr_manager = nil
	}
	if warpserv.mgr_service != nil {
		warpserv.mgr_service.Close()
		warpserv.mgr_service = nil
	}
	return nil
}

// Ensure that the Warp service is set to startup automatically
func (warpserv *WarpServ) EnsureIsEnabled() error {
	mgr_config, err := warpserv.mgr_service.Config()
	if err != nil {
		return fmt.Errorf("failed to get service config:\n %w", err)
	}

	enabled, err := warpserv.IsEnabled()
	if err != nil {
		return err
	}
	if enabled {
		return nil
	} else {
		newConfig := mgr.Config{
			StartType: mgr.StartAutomatic,
			// Keep other settings the same
			DisplayName:      mgr_config.DisplayName,
			Description:      mgr_config.Description,
			BinaryPathName:   mgr_config.BinaryPathName,
			LoadOrderGroup:   mgr_config.LoadOrderGroup,
			Dependencies:     mgr_config.Dependencies,
			ServiceStartName: mgr_config.ServiceStartName,
			DelayedAutoStart: mgr_config.DelayedAutoStart,
			ErrorControl:     mgr_config.ErrorControl,
			ServiceType:      mgr_config.ServiceType,
		}

		err := warpserv.mgr_service.UpdateConfig(newConfig)
		if err != nil {
			return fmt.Errorf("failed to update service config: %w", err)
		}

		return warp_serv.waitForWarpServToUpdate(20, 500*time.Millisecond)
	}
}

// Check if the Warp service is set to startup automatically
func (warpserv *WarpServ) IsEnabled() (bool, error) {
	mgr_config, err := warpserv.mgr_service.Config()
	if err != nil {
		return false, fmt.Errorf("failed to get service config:\n %w", err)
	}

	if mgr_config.StartType == mgr.StartAutomatic {
		return true, nil
	} else {
		return false, nil
	}
}

// Wait for Warp process to start
func (warpserv *WarpServ) waitForWarpServToUpdate(maxAttempts int, waitTime time.Duration) error {
	enabled, err := warpserv.IsEnabled()
	if err != nil {
		return err
	}
	if enabled {
		return nil
	}

	log, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	for attempt := 1; attempt < maxAttempts; attempt++ {
		log.Debug().Msgf("[%v/%v] Waiting for '%v' to start...",
			attempt, maxAttempts, warpserv.appconf.WarpServiceName)
		time.Sleep(waitTime)

		enabled, err = warpserv.IsEnabled()
		if err != nil {
			return err
		}
		if enabled {
			log.Info().Msgf("'%v' successfully started after %v attempts",
				warpserv.appconf.WarpServiceName, attempt)
			return nil
		}
	}

	return fmt.Errorf("could not enable '%v' for startup after %v attempts",
		warpserv.appconf.WarpServiceName, maxAttempts)
}
