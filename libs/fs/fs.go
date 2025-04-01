package fs

import (
	"fmt"
	"os"
)

func FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // File doesn't exist
		}
		return false, fmt.Errorf("Could not find the '%v' file:\n %w", path, err)
	}
	return !info.IsDir(), nil
}

func DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Directory doesn't exist
		}
		return false, fmt.Errorf("Could not find the '%v' folder:\n %w", path, err)
	}
	return info.IsDir(), nil
}
