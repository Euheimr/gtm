package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gtm"
	"gtm/ui"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"time"
)

var fMain *tview.Flex

func init() {
	// Logging will not work as expected unless we set it first before ANYTHING
	gtm.SetupFileLogging()

	if gtm.Cfg.Debug {
		// This is for performance profiling (pprof). Open a web browser and see:
		//	http://localhost:6060/debug/pprof/
		go func() {
			// For docs, see: https://pkg.go.dev/runtime/pprof and:
			//	https://github.com/google/pprof/blob/main/doc/README.md
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	// Seed the initial values & data before setting up the rest of the app
	gtm.GetCPUInfo()
	gtm.GetDiskInfo()
	gtm.GetGPUInfo()
	gtm.GetHostInfo()
	gtm.GetMemoryInfo()
	gtm.GetNetworkInfo()
}

func main() {
	// Scaffold the FlexBox `Main` and layout
	fMain = ui.SetupLayout()

	// Create a new application and be sure to set the root object
	app := tview.NewApplication()
	// TODO: Enable mouse clicking when adding mouse input captures in the future
	app.SetRoot(fMain, true).EnableMouse(false)

	// Setup keybinds ...
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			app.Stop()
		default:
			return event
		}
		return event
	})

	// Setup goroutines handling the drawing of each box here
	slog.Info("Setting up UI goroutines ...")
	go ui.UpdateCPU(app, false, gtm.Cfg.UpdateInterval)
	go ui.UpdateCPUTemp(app, true, gtm.Cfg.UpdateInterval)
	go ui.UpdateDisk(app, true, gtm.Cfg.UpdateInterval)
	if gtm.HasGPU() {
		slog.Info("GPU detected! Setting up GPU/GPUTemp UI goroutines ...")
		go ui.UpdateGPU(app, true, gtm.Cfg.UpdateInterval)
		go ui.UpdateGPUTemp(app, true, gtm.Cfg.UpdateInterval)
	}
	go ui.UpdateMemory(app, true, gtm.Cfg.UpdateInterval)
	go ui.UpdateNetwork(app, true, gtm.Cfg.UpdateInterval)
	go ui.UpdateProcesses(app, true, gtm.Cfg.UpdateInterval)

	slog.Info("Waiting for goroutines to start up ...")
	time.Sleep(40 * time.Millisecond) // wait to start up all the goroutines

	// START APP
	slog.Info("Starting the app ...")
	if err := app.Run(); err != nil {
		slog.Error("Failed to run the app! " + err.Error())
		panic(err)
	}
}
