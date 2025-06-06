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

type Leveler interface {
	Level() slog.Level
}

const LevelPerf = slog.Level(-5)

var LevelNames = map[slog.Leveler]string{
	LevelPerf:       "PERF",
	slog.LevelDebug: strings.ToUpper(slog.LevelDebug.String()),
	slog.LevelInfo:  strings.ToUpper(slog.LevelInfo.String()),
	slog.LevelWarn:  strings.ToUpper(slog.LevelWarn.String()),
	slog.LevelError: strings.ToUpper(slog.LevelError.String()),
}

func SetupFileLogging() {
	var (
		file io.Writer
		opts *slog.HandlerOptions
	)
	opts = &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelName, exists := LevelNames[level]
				if !exists {
					levelName = level.String()
				}
				a.Value = slog.StringValue(levelName)
			}
			return a
		},
	}

	if Cfg.TraceFunctionLogging {
		opts.AddSource = true
	} else {
		opts.AddSource = false
	}

	if Cfg.Debug && Cfg.PerformanceLogging {
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
				a.Value = slog.StringValue(LevelNames[slog.LevelDebug])
			case LevelPerf:
				a.Value = slog.StringValue(LevelNames[LevelPerf])
			case slog.LevelInfo:
				a.Value = slog.StringValue(LevelNames[slog.LevelInfo])
			case slog.LevelWarn:
				a.Value = slog.StringValue(LevelNames[slog.LevelWarn])
			case slog.LevelError:
				a.Value = slog.StringValue(LevelNames[slog.LevelError])
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

	logFilepath := filepath.Join(logsDir, timestampString+"_"+LevelNames[opts.Level]+".log")

	if file, err = os.Create(logFilepath); err != nil {
		slog.Error("Failed to create log file at " + logFilepath + " !")
	}

	fileHandler := slog.NewJSONHandler(file, opts)
	fileLogger := slog.New(fileHandler)
	slog.SetDefault(fileLogger)
}
