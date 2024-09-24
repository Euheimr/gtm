package internal

import (
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type ConfigVars struct {
	Debug          bool
	Production     bool
	ClearOldLogs   bool
	UpdateInterval time.Duration
	Celsius        bool
	EnableGPU      bool
}

var Cfg ConfigVars

func init() {
	readConfig()
}

func readConfig() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file")
	}

	DEBUG, err := strconv.ParseBool(os.Getenv("DEBUG"))
	PRODUCTION, err := strconv.ParseBool(os.Getenv("PRODUCTION"))
	CLEAR_OLD_LOGS, err := strconv.ParseBool(os.Getenv("CLEAR_OLD_LOGS"))
	UPDATE_INTERVAL, err := strconv.ParseInt(os.Getenv("UPDATE_INTERVAL"), 10, 32)
	CELSIUS, err := strconv.ParseBool(os.Getenv("CELSIUS"))
	ENABLE_GPU, err := strconv.ParseBool(os.Getenv("ENABLE_GPU"))

	Cfg = ConfigVars{
		Debug:          DEBUG,
		Production:     PRODUCTION,
		ClearOldLogs:   CLEAR_OLD_LOGS,
		UpdateInterval: time.Duration(UPDATE_INTERVAL) * time.Millisecond,
		Celsius:        CELSIUS,
		EnableGPU:      ENABLE_GPU,
	}

	if Cfg.Debug {
		slog.Debug("DEBUG=" + os.Getenv("DEBUG"))
		slog.Debug("PRODUCTION=" + os.Getenv("PRODUCTION"))
		slog.Debug("CLEAR_OLD_LOGS=" + os.Getenv("CLEAR_OLD_LOGS"))
		slog.Debug("UPDATE_INTERVAL=" + os.Getenv("UPDATE_INTERVAL") + "ms")
		slog.Debug("CELSIUS=" + os.Getenv("CELSIUS"))
		slog.Debug("ENABLE_GPU=" + os.Getenv("ENABLE_GPU"))
	}
}
