package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "MyGoService"
const serviceDisplayName = "My Go Service"
const serviceDescription = "A simple Go service that starts automatically"

type myService struct{}

// Execute implements the service logic
func (m *myService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	changes <- svc.Status{State: svc.StartPending}

	// Set up logging
	logFile, err := os.OpenFile(filepath.Join(filepath.Dir(os.Args[0]), "service.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	log.Println("Service started")

	// Service is now running
	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	// Main service loop
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			log.Println("Service is running - heartbeat")
			// Your service work here

		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				log.Println("Service stopping")
				changes <- svc.Status{State: svc.StopPending}
				break loop
			default:
				log.Printf("Unexpected control request: %d", c)
			}
		}
	}

	return false, 0
}

func runService() {
	err := svc.Run(serviceName, &myService{})
	if err != nil {
		log.Fatalf("Service failed: %v", err)
	}
}

func installService() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %v", err)
	}

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("could not connect to service manager: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", serviceName)
	}

	s, err = m.CreateService(
		serviceName,
		exePath,
		mgr.Config{
			DisplayName: serviceDisplayName,
			Description: serviceDescription,
			StartType:   mgr.StartAutomatic, // Set to start automatically
		})
	if err != nil {
		return fmt.Errorf("could not create service: %v", err)
	}
	defer s.Close()

	// Set up event logging
	err = eventlog.InstallAsEventCreate(serviceName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("could not set up event logging: %v", err)
	}

	return nil
}

func removeService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("could not connect to service manager: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed", serviceName)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return fmt.Errorf("could not remove service: %v", err)
	}

	err = eventlog.Remove(serviceName)
	if err != nil {
		return fmt.Errorf("could not remove event log: %v", err)
	}

	return nil
}

func startService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("could not connect to service manager: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("could not open service: %v", err)
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}

	return nil
}

func stopService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("could not connect to service manager: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("could not open service: %v", err)
	}
	defer s.Close()

	_, err = s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("could not stop service: %v", err)
	}

	return nil
}

func usage() {
	fmt.Printf("Usage:\n")
	fmt.Printf("  %s install   - Install the service\n", os.Args[0])
	fmt.Printf("  %s remove    - Remove the service\n", os.Args[0])
	fmt.Printf("  %s start     - Start the service\n", os.Args[0])
	fmt.Printf("  %s stop      - Stop the service\n", os.Args[0])
	fmt.Printf("  %s debug     - Run service in debug mode\n", os.Args[0])
}

// func main() {
// 	isService, err := svc.IsWindowsService()
// 	if err != nil {
// 		log.Fatalf("Failed to determine if we're running as Windows Service: %v", err)
// 	}

// 	// If no command line arguments and not running as service, run as service
// 	if isService && len(os.Args) == 1 {
// 		runService()
// 		return
// 	}

// 	// Ensure to run myself as admin
// 	if err := admin.EnsureAdmin(); err != nil {
// 		log.Fatalf("Could not ensure if I ran as admin:\n %v", err)
// 	}

// 	// If running as service or with arguments, handle commands
// 	if len(os.Args) < 2 {
// 		usage()
// 		return
// 	}

// 	cmd := os.Args[1]
// 	switch cmd {
// 	case "install":
// 		err := installService()
// 		if err != nil {
// 			log.Fatalf("Failed to install service: %v", err)
// 		}
// 	case "remove":
// 		err := removeService()
// 		if err != nil {
// 			log.Fatalf("Failed to remove service: %v", err)
// 		}
// 	case "start":
// 		err := startService()
// 		if err != nil {
// 			log.Fatalf("Failed to start service: %v", err)
// 		}
// 	case "stop":
// 		err := stopService()
// 		if err != nil {
// 			log.Fatalf("Failed to stop service: %v", err)
// 		}
// 	case "debug":
// 		// Run service in debug mode
// 		debug.Run(serviceName, &myService{})
// 		return
// 	default:
// 		usage()
// 		return
// 	}

// 	fmt.Printf("Command '%s' completed successfully\n", cmd)
// }
