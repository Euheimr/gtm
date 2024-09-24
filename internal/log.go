package internal

/// This article was a nice guide to colorized terminal logging:
// https://dusted.codes/creating-a-pretty-console-logger-using-gos-slog-package

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Handler struct {
	bytes   *bytes.Buffer
	handler slog.Handler
	mutex   *sync.Mutex
}

const (
	reset = "\033[0m"
	//black        = 30
	//red          = 31
	//green        = 32
	//yellow       = 33
	//blue         = 34
	//magenta      = 35
	cyan      = 36
	lightGray = 37
	darkGray  = 90
	lightRed  = 91
	//lightGreen   = 92
	lightYellow = 93
	//lightBlue    = 94
	//lightMagenta = 95
	//lightCyan    = 96
	white = 97
)

const (
	timeFormat = "[15:04:05.000]"
)

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}

func (h *Handler) computeAttrs(ctx context.Context, record slog.Record) (map[string]any, error) {
	h.mutex.Lock()

	// Reset the buffer and release the mutex when everything is complete
	defer func() {
		h.bytes.Reset()
		h.mutex.Unlock()
	}()

	if err := h.handler.Handle(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to call inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.bytes.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	return attrs, nil
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = colorize(darkGray, level)
	case slog.LevelInfo:
		level = colorize(cyan, level)
	case slog.LevelWarn:
		level = colorize(lightYellow, level)
	case slog.LevelError:
		level = colorize(lightRed, level)
	}

	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(attrs, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON attrs: %w", err)
	}

	fmt.Println(
		colorize(lightGray, r.Time.Format(timeFormat)),
		level,
		colorize(white, r.Message),
		colorize(darkGray, string(b)),
	)

	return nil
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		bytes:   h.bytes,
		handler: h.handler.WithAttrs(attrs),
		mutex:   h.mutex,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		bytes:   h.bytes,
		handler: h.handler.WithGroup(name),
		mutex:   h.mutex,
	}
}

func suppressDefaults(
	next func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {

	return func(groups []string, a slog.Attr) slog.Attr {
		// Since our handler already handles Time, Level and Message formatting, we need
		//	to filter out these three slog attributes
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}

		if next == nil {
			return a
		}

		return next(groups, a)
	}
}

func NewHandler(opts *slog.HandlerOptions) *Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	b := &bytes.Buffer{}
	return &Handler{
		bytes: b,
		handler: slog.NewJSONHandler(b,
			&slog.HandlerOptions{
				AddSource: opts.AddSource,
				Level:     opts.Level,
				// we want to keep the ability to state what attributes we want to replace
				//	when logging, so we just filter out Time, Level and Message (because
				//	these are colorized) but keep all other attributes that we want to
				//	remove or replace if stated
				ReplaceAttr: suppressDefaults(opts.ReplaceAttr),
			}),
		mutex: &sync.Mutex{},
	}
}

//func SetupLogging() {
//	var opts *slog.HandlerOptions
//
//	if Cfg.Production && Cfg.Debug {
//		opts = &slog.HandlerOptions{
//			AddSource: true,
//			Level:     slog.LevelInfo,
//		}
//	} else if Cfg.Debug {
//		opts = &slog.HandlerOptions{
//			AddSource: true,
//			Level:     slog.LevelDebug,
//		}
//	} else if Cfg.Production {
//		opts = &slog.HandlerOptions{
//			AddSource: true,
//			Level:     slog.LevelWarn,
//		}
//	}
//
//	//consoleLogger := slog.New(NewHandler(opts))
//
//	fHandler := slog.NewJSONHandler(os.Stdout, opts)
//	fileLogger := slog.New(fHandler)
//
//	slog.SetDefault(fileLogger)
//	//slog.SetLogLoggerLevel(slog.LevelDebug)
//
//	slog.Debug("Initialized logging")
//	//slog.Debug("debug test")
//	//slog.Info("info test")
//	//slog.Warn("warn test")
//	//slog.Error("error test")
//}

func SetupFileLogging() {
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("Failed to get current working directory !")
	}
	cwd = cwd + "\\"
	logsDir := cwd + "log\\"

	if Cfg.ClearOldLogs {
		err = os.RemoveAll(logsDir + "\\")
		if err != nil {
			slog.Error("Failed to remove old log files !")
		}
	}

	err = os.Mkdir(logsDir, 0750)
	if err != nil && os.IsNotExist(err) {
		slog.Error("Failed to create directory: " + logsDir + " !")
	}

	err = os.Chdir(logsDir)
	if err != nil {
		slog.Error("Failed to change directory: " + logsDir)
	}

	timestamp := time.Now().Format(time.DateTime)
	timestampString := strings.ReplaceAll(timestamp, ":", ".")
	timestampString = strings.ReplaceAll(timestampString, " ", "_")

	logPathFilename := logsDir + timestampString + "_test.log"
	file, err := os.Create(logPathFilename)
	if err != nil {
		slog.Error("Failed to create log file at " + logPathFilename + " !")
	}

	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
		//ReplaceAttr: nil,
	}

	fileHandler := slog.NewJSONHandler(file, opts)
	fileLogger := slog.New(fileHandler)
	slog.SetDefault(fileLogger)
}
