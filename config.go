package gtm

import (
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type ConfigVars struct {
	Debug                bool
	Production           bool
	DeleteOldLogs        bool
	TraceFunctionLogging bool
	UpdateInterval       time.Duration
	Celsius              bool
	EnableGPU            bool
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
	DELETE_OLD_LOGS, err := strconv.ParseBool(os.Getenv("DELETE_OLD_LOGS"))
	TRACE_FUNCTION_LOGGING, err := strconv.ParseBool(os.Getenv("TRACE_FUNCTION_LOGGING"))
	UPDATE_INTERVAL, err := strconv.ParseInt(os.Getenv("UPDATE_INTERVAL"), 10, 64)
	CELSIUS, err := strconv.ParseBool(os.Getenv("CELSIUS"))

	Cfg = ConfigVars{
		Debug:                DEBUG,
		DeleteOldLogs:        DELETE_OLD_LOGS,
		TraceFunctionLogging: TRACE_FUNCTION_LOGGING,
		UpdateInterval:       time.Duration(UPDATE_INTERVAL) * time.Millisecond,
		Celsius:              CELSIUS,
	}

	if Cfg.Debug {
		slog.Debug("DELETE_OLD_LOGS=" + os.Getenv("DELETE_OLD_LOGS"))
		slog.Debug("TRACE_FUNCTION_LOGGING=" + os.Getenv("TRACE_FUNCTION_LOGGING"))
		slog.Debug("UPDATE_INTERVAL=" + os.Getenv("UPDATE_INTERVAL") + "ms")
		slog.Debug("CELSIUS=" + os.Getenv("CELSIUS"))
	}
}
