package gtm

import (
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const LevelPerf = slog.Level(-5)

func SetupFileLogging() {
	var (
		file        io.Writer
		opts        *slog.HandlerOptions
		logLevelStr string
	)
	opts = &slog.HandlerOptions{}

	if Cfg.TraceFunctionLogging {
		opts.AddSource = true
	} else {
		opts.AddSource = false
	}

	if Cfg.Debug && Cfg.PerformanceTest {
		opts.Level = LevelPerf
	} else if Cfg.Debug {
		opts.Level = slog.LevelDebug
	} else {
		opts.Level = slog.LevelInfo
	}

	opts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.LevelKey {
			level := a.Value.Any().(slog.Level)

			switch level {
			case slog.LevelDebug:
				a.Value = slog.StringValue("DEBUG")
			case LevelPerf:
				a.Value = slog.StringValue("PERF")
			case slog.LevelInfo:
				a.Value = slog.StringValue("INFO")
			case slog.LevelWarn:
				a.Value = slog.StringValue("WARN")
			case slog.LevelError:
				a.Value = slog.StringValue("ERROR")
			}
		}
		return a
	}

	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("Failed to get current working directory !")
	}
	logsDir := filepath.Join(cwd, "log")

	if Cfg.DeleteOldLogs {
		if err = os.RemoveAll(logsDir); err != nil {
			slog.Error("Failed to remove old log files !")
		}
	}

	err = os.Mkdir(logsDir, 0750)
	if errors.Is(err, fs.ErrExist) {
		slog.Debug("Log directory exists")
	} else if errors.Is(err, fs.ErrNotExist) {
		slog.Error("Failed to create directory: " + logsDir + " !")
	}

	if err = os.Chdir(logsDir); err != nil {
		slog.Error("Failed to change directory: " + logsDir)
	}

	timestamp := time.Now().Format(time.DateTime)
	timestampString := strings.ReplaceAll(timestamp, ":", ".")
	timestampString = strings.ReplaceAll(timestampString, " ", "_")

	switch opts.Level {
	case LevelPerf:
		logLevelStr = "perf"
	case slog.LevelDebug:
		logLevelStr = "debug"
	case slog.LevelInfo:
		logLevelStr = "info"
	case slog.LevelWarn:
		logLevelStr = "warn"
	case slog.LevelError:
		logLevelStr = "error"
	default:
		logLevelStr = "info"
	}
	logFilepath := filepath.Join(logsDir, timestampString+"_"+logLevelStr+".log")

	if file, err = os.Create(logFilepath); err != nil {
		slog.Error("Failed to create log file at " + logFilepath + " !")
	}

	fileHandler := slog.NewJSONHandler(file, opts)
	fileLogger := slog.New(fileHandler)
	slog.SetDefault(fileLogger)
}
