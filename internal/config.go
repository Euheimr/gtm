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
	UPDATE_INTERVAL, err := strconv.ParseInt(os.Getenv("UPDATE_INTERVAL"), 10, 32)
	CELSIUS, err := strconv.ParseBool(os.Getenv("CELSIUS"))
	ENABLE_GPU, err := strconv.ParseBool(os.Getenv("ENABLE_GPU"))

	Cfg = ConfigVars{
		Debug:          DEBUG,
		UpdateInterval: time.Duration(UPDATE_INTERVAL) * time.Millisecond,
		Celsius:        CELSIUS,
		EnableGPU:      ENABLE_GPU,
	}

	if Cfg.Debug {
		slog.Debug("DEBUG=" + os.Getenv("DEBUG"))
		slog.Debug("UPDATE_INTERVAL=%dms\n", os.Getenv("UPDATE_INTERVAL"))
		slog.Debug("CELSIUS=%v\n", os.Getenv("CELSIUS"))
		slog.Debug("ENABLE_GPU=%v\n", os.Getenv("ENABLE_GPU"))
	}
}
