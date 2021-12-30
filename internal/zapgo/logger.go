package zapgo

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
	logrusadapter "logur.dev/adapter/logrus"
	"logur.dev/logur"
)

type LogConfig struct {
	// Format specifies the output log format.
	// Accepted values are: json, logfmt
	Format string
	// Level is the minimum log level that should appear on the output.
	Level string
	// NoColor makes sure that no log output gets colorized.
	NoColor bool
}

type Logger logur.LoggerFacade

// NewLogger creates a new logger.
func NewLogger(config LogConfig) logur.LoggerFacade {
	logger := logrus.New()

	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:             config.NoColor,
		EnvironmentOverrideColors: true,
	})

	switch config.Format {
	case "logfmt":
		// Already the default

	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	if level, err := logrus.ParseLevel(config.Level); err == nil {
		logger.SetLevel(level)
	}

	return logrusadapter.New(logger)
}

// SetStandardLogger sets the global logger's output to a custom logger instance.
func SetStandardLogger(logger logur.Logger) {
	log.SetOutput(logur.NewLevelWriter(logger, logur.Info))
}
