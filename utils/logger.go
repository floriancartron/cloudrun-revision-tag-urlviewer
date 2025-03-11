package utils

import (
	"log/slog"
	"os"
)

// String to loglevel matrix
var litteralLogLevels = map[string]slog.Level{
	"DEBUG": slog.LevelDebug,
	"INFO":  slog.LevelInfo,
	"WARN":  slog.LevelWarn,
	"ERROR": slog.LevelError,
}

// Return custom *slog.Logger, without timestamp and with desired level
func NewLogger(logLevel string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: litteralLogLevels[logLevel],
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove time from the output as it is injected by google cloud.
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			if a.Key == slog.LevelKey {
				return slog.Attr{Key: "severity", Value: a.Value} // Change "level" to "severity" for google cloud logging
			}
			return a
		},
	}))
}
