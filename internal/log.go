package internal

import (
	"log/slog"
	"os"
)

func SetupLogging() {
	var opts *slog.HandlerOptions

	if Cfg.Debug {
		opts = &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}
	} else {
		opts = &slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelError,
		}
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)

	logger := slog.New(handler)

	slog.SetDefault(logger)
	//slog.SetLogLoggerLevel(slog.LevelDebug)

	slog.Debug("Initialized logging")

}
