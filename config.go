package gtm

import (
	"github.com/joho/godotenv"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ConfigVars struct {
	Celsius              bool
	DeleteOldLogs        bool
	Debug                bool
	PerformanceLogging   bool
	PerformanceLoggingUI bool
	TraceFunctionLogging bool
	UpdateInterval       time.Duration
}

var CFG_DEFAULT = ConfigVars{
	Celsius:              true,
	DeleteOldLogs:        false,
	Debug:                false,
	PerformanceLogging:   false,
	PerformanceLoggingUI: false,
	TraceFunctionLogging: false,
	UpdateInterval:       500 * time.Millisecond,
}

const CONFIG_FILENAME = ".env"

var Cfg ConfigVars

func getRootDir() (dir string, err error) {
	if dir, err = os.Getwd(); err != nil {
		slog.Error("Failed to get current working directory: " + err.Error())
		return "", err
	}

	for {
		if lastDir := strings.Split(dir, "\\"); lastDir[len(lastDir)-1] != "gtm" {
			dir = filepath.Dir(dir)
		} else {
			break
		}
	}

	return dir, nil
}

func writeConfig() (err error) {
	var (
		dir     string
		cfgPath string
	)

	if dir, err = getRootDir(); err != nil {
		log.Println("Failed to get config file path: " + err.Error())
		return err
	}

	cfgPath = filepath.Join(dir, CONFIG_FILENAME)

	if _, err = os.Stat(cfgPath); os.IsExist(err) {
		log.Println(CONFIG_FILENAME + " file exists! " + err.Error())
		return err
	}

	cfgData := []byte("DEBUG=" + strconv.FormatBool(CFG_DEFAULT.Debug) + "\n\n" +
		"# Temperature units - `true` for Celsius, `false` for Fahrenheit\n" +
		"CELSIUS=" + strconv.FormatBool(CFG_DEFAULT.Celsius) + "\n" +
		"# Performance logging is quite heavy on disk read/writes\n" +
		"# !!! IMPORTANT !!! This option ONLY works if DEBUG is also true\n" +
		"PERFORMANCE_LOGGING=" + strconv.FormatBool(CFG_DEFAULT.PerformanceLogging) + "\n" +
		"PERFORMANCE_LOGGING_UI=" + strconv.FormatBool(CFG_DEFAULT.PerformanceLoggingUI) + "\n\n" +
		"# Set how frequently to update the UI (in milliseconds - 1000ms equals 1 second)\n" +
		"UPDATE_INTERVAL=" + strings.TrimRight(CFG_DEFAULT.UpdateInterval.String(), "ms") + "\n\n" +
		"### Logging\n" +
		"DELETE_OLD_LOGS=" + strconv.FormatBool(CFG_DEFAULT.DeleteOldLogs) + "\n" +
		"TRACE_FUNCTION_LOGGING=" + strconv.FormatBool(CFG_DEFAULT.TraceFunctionLogging) + "\n\n" +
		"# Enable or disable grouping of processes in Processes table (true or false)\n" +
		"# GROUP_PROCESSES=true" + "\n")

	if err = os.WriteFile(cfgPath, cfgData, 0644); err != nil {
		log.Println("Failed to write config file: " + err.Error())
		return err
	}

	if _, err = os.Stat(cfgPath); os.IsExist(err) {
		log.Println("Wrote config file to " + dir + CONFIG_FILENAME)
	}

	return nil
}

func ReadConfig() (err error) {
	var (
		celsius              bool
		deleteOldLogs        bool
		debug                bool
		performanceLogging   bool
		performanceLoggingUI bool
		traceFunctionLogging bool
		updateInterval       int64
	)

	// seed the default values first, then override those defaults with values read
	//	from the config file (.env)
	Cfg = CFG_DEFAULT

	rootDir, err := getRootDir()
	if err != nil {
		log.Println("Failed to get root directory: " + err.Error())
		return err
	}

	err = godotenv.Load(filepath.Join(rootDir, CONFIG_FILENAME))
	if err != nil {
		log.Println("Failed to read config vars from `.env` ... using defaults")
		if err = writeConfig(); err != nil {
			log.Println(err.Error())
		}
	} else {
		// Reading .env was successful ... populate the values from .env file

		if celsius, err = strconv.ParseBool(os.Getenv("CELSIUS")); err == nil {
			Cfg.Celsius = celsius
		} else {
			log.Println("Failed to parse boolean: CELSIUS ... " +
				"using default value: " + strconv.FormatBool(CFG_DEFAULT.Celsius))
		}

		if deleteOldLogs, err = strconv.ParseBool(os.Getenv("DELETE_OLD_LOGS")); err == nil {
			Cfg.DeleteOldLogs = deleteOldLogs
		} else {
			log.Println("Failed to parse boolean: deleteOldLogs ... " +
				"using default value: " + strconv.FormatBool(CFG_DEFAULT.DeleteOldLogs))
		}

		if debug, err = strconv.ParseBool(os.Getenv("DEBUG")); err == nil {
			Cfg.Debug = debug
		} else {
			log.Println("Failed to parse boolean: DEBUG ... using default value: " +
				strconv.FormatBool(CFG_DEFAULT.Debug))
		}

		if performanceLogging, err = strconv.ParseBool(os.Getenv("PERFORMANCE_LOGGING")); err == nil {
			Cfg.PerformanceLogging = performanceLogging
		} else {
			log.Println("Failed to parse boolean: PERFORMANCE_LOGGING ... using default: " +
				strconv.FormatBool(CFG_DEFAULT.PerformanceLogging))
		}

		if performanceLoggingUI, err = strconv.ParseBool(os.Getenv("PERFORMANCE_LOGGING_UI")); err == nil {
			Cfg.PerformanceLoggingUI = performanceLoggingUI
		} else {
			log.Println("Failed to parse boolean: PERFORMANCE_LOGGING_UI ... using default: " +
				strconv.FormatBool(CFG_DEFAULT.PerformanceLoggingUI))
		}

		if traceFunctionLogging, err = strconv.ParseBool(os.Getenv("TRACE_FUNCTION_LOGGING")); err == nil {
			Cfg.TraceFunctionLogging = traceFunctionLogging
		} else {
			log.Println("Failed to parse boolean: traceFunctionLogging ... using default: " +
				strconv.FormatBool(CFG_DEFAULT.TraceFunctionLogging))
		}

		if updateInterval, err = strconv.ParseInt(os.Getenv("UPDATE_INTERVAL"), 10, 64); err == nil {
			Cfg.UpdateInterval = time.Duration(updateInterval) * time.Millisecond
		} else {
			log.Println("Failed to parse integer: UPDATE_INTERVAL ... using default: " +
				CFG_DEFAULT.UpdateInterval.String())
		}
	}

	return nil
}
