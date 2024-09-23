package internal

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
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
		log.Fatal("Error loading .env file")
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
		fmt.Printf("DEBUG=%v\n", Cfg.Debug)
		fmt.Printf("UPDATE_INTERVAL=%dms\n", Cfg.UpdateInterval/time.Millisecond)
		fmt.Printf("CELSIUS=%v\n", Cfg.Celsius)
		fmt.Printf("ENABLE_GPU=%v\n", Cfg.EnableGPU)
	}
}
