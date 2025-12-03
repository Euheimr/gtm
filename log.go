package gtm

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
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

func createLogDir(dir string) error {
	err := os.Mkdir(dir, 0o750)
	if errors.Is(err, fs.ErrExist) {
		slog.Debug("Log directory exists")
	} else {
		return fmt.Errorf("failed to create directory: %s", dir)
	}

	return nil
}

func deleteLogs(dir string) error {
	if Cfg.DeleteOldLogs {
		files, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("problem occurred reading directory '%s': %s", dir, err.Error())
		}

		for _, file := range files {
			splitFilename := strings.Split(file.Name(), ".")
			fileType := splitFilename[len(splitFilename)-1]

			// Precautionary measure to ensure we only delete files that end in .log and not
			// deleting directories named 'log'
			if !file.IsDir() && fileType == "log" {
				fp := filepath.Join(dir, file.Name())
				if err = os.Remove(fp); err != nil {
					return fmt.Errorf("failed to delete log file: %s", file.Name())
				} else {
					log.Printf("Successfully deleted log file: '%s' at '%s'", file.Name(), dir)
				}
			}
		}
	}

	return nil
}

func createLogFile(dir string, level slog.Leveler) (*os.File, error) {
	timestamp := time.Now().Format(time.DateTime)
	timestampString := strings.ReplaceAll(timestamp, ":", ".")
	timestampString = strings.ReplaceAll(timestampString, " ", "_")

	logFilepath := filepath.Join(dir, timestampString+"_"+LevelNames[level]+".log")

	file, err := os.Create(filepath.Clean(logFilepath))
	if err != nil {
		return nil, fmt.Errorf("failed to create log file at %s", logFilepath)
	}

	return file, nil
}

func SetupFileLogging() {
	var (
		file io.Writer
		opts *slog.HandlerOptions
	)

	log.Println("Setup file logging ...")

	opts = &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.LevelKey {
				level, ok := attr.Value.Any().(slog.Level)
				if !ok {
					log.Fatal("attr.Value.Any() is not of type slog.Level !")
				}

				switch level {
				case LevelPerf:
					attr.Value = slog.StringValue(LevelNames[LevelPerf])
				case slog.LevelDebug:
					attr.Value = slog.StringValue(LevelNames[slog.LevelDebug])
				case slog.LevelInfo:
					attr.Value = slog.StringValue(LevelNames[slog.LevelInfo])
				case slog.LevelWarn:
					attr.Value = slog.StringValue(LevelNames[slog.LevelWarn])
				case slog.LevelError:
					attr.Value = slog.StringValue(LevelNames[slog.LevelError])
				}
			}

			return attr
		},
	}

	if Cfg.PerformanceLogging {
		opts.Level = LevelPerf
	} else if Cfg.Debug {
		opts.Level = slog.LevelDebug
	}

	if Cfg.TraceFunctionLogging {
		opts.AddSource = true
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get current working directory !")
	}

	logsDir := filepath.Join(cwd, "log")
	if err = createLogDir(logsDir); err != nil {
		log.Fatal(err.Error())
	}

	if err = os.Chdir(logsDir); err != nil {
		log.Fatalf("Failed to change directory: %s", logsDir)
	}

	if err = deleteLogs(logsDir); err != nil {
		log.Fatal(err.Error())
	}

	file, err = createLogFile(logsDir, opts.Level)
	if err != nil {
		log.Fatal(err.Error())
	}

	fileHandler := slog.NewJSONHandler(file, opts)
	fileLogger := slog.New(fileHandler)

	slog.SetDefault(fileLogger)
	slog.Debug("File logging initialized")
}
