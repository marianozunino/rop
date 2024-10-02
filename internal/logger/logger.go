package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func ConfigureLogger(debug bool) {
	// Set the global logging level based on the debug flag
	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure the logger with optional caller information in debug mode
	logger := log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "", // Empty TimeFormat for default formatting
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		},
		FormatLevel: func(i interface{}) string {
			if debug {
				return strings.ToUpper(fmt.Sprintf("[%s]", i))
			}
			return ""
		},
		FormatCaller: func(i interface{}) string {
			if i == nil {
				return ""
			}
			return filepath.Base(fmt.Sprintf("%s >", i))
		},
		PartsExclude: []string{"time"},
	})

	// Enable caller information if in debug mode
	if debug {
		logger = logger.With().Caller().Logger()
	}

	log.Logger = logger
}
