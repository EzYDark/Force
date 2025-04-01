package admin

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/ezydark/warpenforcer/libs/logger"
	"golang.org/x/sys/windows"
)

func EnsureAdmin() error {
	if !IsSelfAdmin() {
		if err := RunSelfAsAdmin(); err != nil {
			return fmt.Errorf("Could not run it self as admin:\n %w", err)
		}

		log, err := logger.Get()
		if err != nil {
			return fmt.Errorf("Could not get Logger:\n %w", err)
		}
		log.Fatal().Msg("Stopping this instance of program... Starting as admin instead")
	}
	return nil
}

func IsSelfAdmin() bool {
	var sid *windows.SID

	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	return err == nil && member
}

func RunSelfAsAdmin() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	verb := "runas"
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(execPath)
	argPtr, _ := syscall.UTF16PtrFromString(args)
	dirPtr, _ := syscall.UTF16PtrFromString("")

	if err = windows.ShellExecute(
		0,
		verbPtr,
		exePtr,
		argPtr,
		dirPtr,
		windows.SW_NORMAL); err != nil {
		return fmt.Errorf("Could not ShellExecute:\n %w", err)
	}

	return nil
}
