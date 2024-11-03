package gtm

import (
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type ConfigVars struct {
	Celsius              bool
	DeleteOldLogs        bool
	Debug                bool
	PerformanceLogging   bool
	TraceFunctionLogging bool
	UpdateInterval       time.Duration
}

var CFG_DEFAULT = ConfigVars{
	Celsius:              true,
	DeleteOldLogs:        false,
	Debug:                false,
	PerformanceLogging:   false,
	TraceFunctionLogging: false,
	UpdateInterval:       500 * time.Millisecond,
}

var Cfg ConfigVars

func ReadConfig() {
	var (
		err                  error
		celsius              bool
		deleteOldLogs        bool
		debug                bool
		performanceLogging   bool
		traceFunctionLogging bool
		updateInterval       int64
	)

	// seed the default values first, then override those defaults with values read
	//	from the config file (.env)
	Cfg = CFG_DEFAULT

	err = godotenv.Load()
	if err != nil {
		slog.Error("Failed to read config vars from `.env` ... using defaults")
	} else {
		// Reading .env was successful ... populate the values from .env file

		celsius, err = strconv.ParseBool(os.Getenv("CELSIUS"))
		if err == nil {
			Cfg.Celsius = celsius
		} else {
			slog.Error("Failed to parse boolean: CELSIUS ... " +
				"using default value: " + strconv.FormatBool(CFG_DEFAULT.Celsius))
		}

		if deleteOldLogs, err = strconv.ParseBool(os.Getenv("DELETE_OLD_LOGS")); err == nil {
			Cfg.DeleteOldLogs = deleteOldLogs
		} else {
			slog.Error("Failed to parse boolean: deleteOldLogs ... " +
				"using default value: " + strconv.FormatBool(CFG_DEFAULT.DeleteOldLogs))
		}

		if debug, err = strconv.ParseBool(os.Getenv("DEBUG")); err == nil {
			Cfg.Debug = debug
		} else {
			slog.Error("Failed to parse boolean: DEBUG ... using default value: " +
				strconv.FormatBool(CFG_DEFAULT.Debug))
		}

		if performanceLogging, err = strconv.ParseBool(os.Getenv("PERFORMANCE_LOGGING")); err == nil {
			Cfg.PerformanceLogging = performanceLogging
		} else {
			slog.Error("Failed to parse boolean: PERFORMANCE_LOGGING ... using default: " +
				strconv.FormatBool(CFG_DEFAULT.PerformanceLogging))
		}

		traceFunctionLogging, err = strconv.ParseBool(os.Getenv("TRACE_FUNCTION_LOGGING"))
		if err == nil {
			Cfg.TraceFunctionLogging = traceFunctionLogging
		} else {
			slog.Error("Failed to parse boolean: traceFunctionLogging ... " +
				"using default value: " + strconv.FormatBool(CFG_DEFAULT.TraceFunctionLogging))
		}

		updateInterval, err = strconv.ParseInt(os.Getenv("UPDATE_INTERVAL"), 10, 64)
		if err == nil {
			Cfg.UpdateInterval = time.Duration(updateInterval) * time.Millisecond
		} else {
			slog.Error("Failed to parse integer: UPDATE_INTERVAL ... " +
				"using default value: " + CFG_DEFAULT.UpdateInterval.String())
		}

	}
}
