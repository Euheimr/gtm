package gtm

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	cfgFilename                = ".env"
	defaultCelsius             = true
	defaultDebug               = false
	defaultDeleteOldLogs       = false
	defaultDetectVirtualDrives = false
	defaultPerfLogging         = false
	defaultPerfLoggingUI       = false
	defaultTraceFuncLogging    = false
	defaultUpdateInterval      = 1000 * time.Millisecond
)

type ConfigVars struct {
	Celsius              bool
	Debug                bool
	DeleteOldLogs        bool
	DetectVirtualDrives  bool
	PerformanceLogging   bool
	PerformanceLoggingUI bool
	TraceFunctionLogging bool
	UpdateInterval       time.Duration
}

var Cfg *ConfigVars

// Seed the default values first, then override those defaults with values read
// from the config file (.env).
func init() {
	Cfg = &ConfigVars{
		Celsius:              defaultCelsius,
		Debug:                defaultDebug,
		DeleteOldLogs:        defaultDeleteOldLogs,
		DetectVirtualDrives:  defaultDetectVirtualDrives,
		PerformanceLogging:   defaultPerfLogging,
		PerformanceLoggingUI: defaultPerfLoggingUI,
		TraceFunctionLogging: defaultTraceFuncLogging,
		UpdateInterval:       defaultUpdateInterval,
	}
}

func getRootDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %s", err.Error())
	}

	if runtime.GOOS == "windows" {
		for {
			if lastDir := strings.Split(dir, "\\"); lastDir[len(lastDir)-1] != "gtm" {
				dir = filepath.Dir(dir)
			} else {
				break
			}
		}
	}

	return dir, nil
}

func writeConfig() error {
	var (
		dir     string
		cfgPath string
		err     error
	)

	if dir, err = getRootDir(); err != nil {
		return fmt.Errorf("failed to get config file path: %s", err.Error())
	}

	cfgPath = filepath.Join(dir, cfgFilename)

	if _, err = os.Stat(cfgPath); os.IsExist(err) {
		return fmt.Errorf(cfgFilename+" file exists! %s", err.Error())
	}

	cfgData := []byte("DEBUG=" + strconv.FormatBool(Cfg.Debug) + "\n\n" +
		"# Temperature units - `true` for Celsius, `false` for Fahrenheit\n" +
		"CELSIUS=" + strconv.FormatBool(Cfg.Celsius) + "\n" +
		"# Show virtual drives in the HDD/SSD area (google drive, network drives, etc)\n" +
		"DETECT_VIRTUAL_DRIVES=" + strconv.FormatBool(Cfg.DetectVirtualDrives) + "\n" +
		"# Performance logging is quite heavy on disk read/writes\n" +
		"# !!! IMPORTANT !!! This option ONLY works if DEBUG is also true\n" +
		"PERFORMANCE_LOGGING=" + strconv.FormatBool(Cfg.PerformanceLogging) + "\n" +
		"PERFORMANCE_LOGGING_UI=" + strconv.FormatBool(Cfg.PerformanceLoggingUI) + "\n\n" +
		"# Set how frequently to update the UI (in milliseconds - 1000ms equals 1 second)\n" +
		"UPDATE_INTERVAL=" + strings.TrimRight(Cfg.UpdateInterval.String(), "ms") + "\n\n" +
		"### Logging\n" +
		"DELETE_OLD_LOGS=" + strconv.FormatBool(Cfg.DeleteOldLogs) + "\n" +
		"TRACE_FUNCTION_LOGGING=" + strconv.FormatBool(Cfg.TraceFunctionLogging) + "\n\n" +
		"# Enable or disable grouping of processes in Processes table (true or false)\n" +
		"# GROUP_PROCESSES=true" + "\n")

	if err = os.WriteFile(cfgPath, cfgData, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %s", err.Error())
	}

	if _, err = os.Stat(cfgPath); os.IsExist(err) {
		log.Println("Wrote config file to " + dir + cfgFilename)
	}

	return nil
}

func populateConfigVars() {
	log.Println("Populating config vars ...")

	celsius, err := strconv.ParseBool(os.Getenv("CELSIUS"))
	if err == nil {
		Cfg.Celsius = celsius
	} else {
		log.Println("Failed to parse boolean: CELSIUS ... " +
			"using default value: " + strconv.FormatBool(Cfg.Celsius))
	}

	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err == nil {
		Cfg.Debug = debug
	} else {
		log.Println("Failed to parse boolean: DEBUG ... using default value: " +
			strconv.FormatBool(Cfg.Debug))
	}

	deleteOldLogs, err := strconv.ParseBool(os.Getenv("DELETE_OLD_LOGS"))
	if err == nil {
		Cfg.DeleteOldLogs = deleteOldLogs
	} else {
		log.Println("Failed to parse boolean: deleteOldLogs ... " +
			"using default value: " + strconv.FormatBool(Cfg.DeleteOldLogs))
	}

	detectVirtualDrives, err := strconv.ParseBool(os.Getenv("DETECT_VIRTUAL_DRIVES"))
	if err == nil {
		Cfg.DetectVirtualDrives = detectVirtualDrives
	} else {
		log.Println("Failed to parse boolean: DETECT_VIRTUAL_DRIVES ... using default value: " +
			strconv.FormatBool(Cfg.DetectVirtualDrives))
	}

	performanceLogging, err := strconv.ParseBool(os.Getenv("PERFORMANCE_LOGGING"))
	if err == nil {
		Cfg.PerformanceLogging = performanceLogging
	} else {
		log.Println("Failed to parse boolean: PERFORMANCE_LOGGING ... using default: " +
			strconv.FormatBool(Cfg.PerformanceLogging))
	}

	performanceLoggingUI, err := strconv.ParseBool(os.Getenv("PERFORMANCE_LOGGING_UI"))
	if err == nil {
		Cfg.PerformanceLoggingUI = performanceLoggingUI
	} else {
		log.Println("Failed to parse boolean: PERFORMANCE_LOGGING_UI ... using default: " +
			strconv.FormatBool(Cfg.PerformanceLoggingUI))
	}

	traceFunctionLogging, err := strconv.ParseBool(os.Getenv("TRACE_FUNCTION_LOGGING"))
	if err == nil {
		Cfg.TraceFunctionLogging = traceFunctionLogging
	} else {
		log.Println("Failed to parse boolean: traceFunctionLogging ... using default: " +
			strconv.FormatBool(Cfg.TraceFunctionLogging))
	}

	updateInterval, err := strconv.ParseInt(os.Getenv("UPDATE_INTERVAL"), 10, 64)
	if err == nil {
		Cfg.UpdateInterval = time.Duration(updateInterval) * time.Millisecond
	} else {
		log.Println("Failed to parse integer: UPDATE_INTERVAL ... using default: " +
			Cfg.UpdateInterval.String())
	}
}

func ReadConfig() error {
	log.Println("Reading config ...")

	rootDir, err := getRootDir()
	if err != nil {
		return fmt.Errorf("failed to get root directory: %s", err.Error())
	} else {
		log.Println("Root dir is: ", rootDir)
	}

	if err = godotenv.Load(filepath.Join(rootDir, cfgFilename)); err != nil {
		log.Println("Failed to read config vars from `.env` ... using defaults")

		if err = writeConfig(); err != nil {
			return fmt.Errorf("%s", err.Error())
		}
	} else {
		// Reading .env was successful ... populate the values from .env file
		populateConfigVars()
	}

	return nil
}

func (c *ConfigVars) SetUpdateInterval(milliseconds time.Duration) {
	c.UpdateInterval = milliseconds
}

func (c *ConfigVars) ResetUpdateInterval() {
	if updateInterval, err := strconv.ParseInt(os.Getenv("UPDATE_INTERVAL"), 10, 64); err == nil {
		c.UpdateInterval = time.Duration(updateInterval) * time.Millisecond
	} else {
		log.Println("Failed to parse integer: UPDATE_INTERVAL ... using default: " +
			defaultUpdateInterval.String() + "ms")
		c.UpdateInterval = defaultUpdateInterval
	}
}
