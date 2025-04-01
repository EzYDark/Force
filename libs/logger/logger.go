package logger

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
)

var logger *zerolog.Logger

func Init() (*zerolog.Logger, error) {
	if logger != nil {
		return nil, errors.New("logger is already initialized")
	}

	consoleOutput := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05.000",
		NoColor:    false,
	}

	consoleOutput.FormatLevel = func(i interface{}) string {
		levelStr := strings.ToUpper(fmt.Sprintf("%s", i))

		switch levelStr {
		case "DEBUG":
			return color.New(color.FgBlue).Sprintf("[%s]", levelStr)
		case "INFO":
			return color.New(color.FgGreen).Sprintf("[%s]", levelStr)
		case "WARN":
			return color.New(color.FgYellow).Sprintf("[%s]", levelStr)
		case "ERROR":
			return color.New(color.FgRed).Sprintf("[%s]", levelStr)
		case "FATAL":
			return color.New(color.FgRed, color.Bold).Sprintf("[%s]", levelStr)
		default:
			return color.New(color.FgWhite).Sprintf("[%s]", levelStr)
		}
	}

	consoleOutput.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}

	consoleOutput.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s=", i)
	}

	consoleOutput.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}

	zerolog := zerolog.New(consoleOutput).With().Timestamp().Logger()
	logger = &zerolog
	return logger, nil
}

func Get() (*zerolog.Logger, error) {
	if logger == nil {
		return nil, errors.New("logger not initialized")
	}
	return logger, nil
}
