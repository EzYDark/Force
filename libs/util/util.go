package util

import (
	"bufio"
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Stop the execution of the program until the user presses Enter.
func WaitForInput() error {
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
