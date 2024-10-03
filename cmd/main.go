package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gtm"
	"gtm/ui"
	"log/slog"
	"runtime"
)

var fMain *tview.Flex

func init() {
	// Logging will not work as expected unless we set it first before ANYTHING
	gtm.SetupFileLogging()

	gtm.GetHostInfo()
	gtm.GetGPUInfo()
}

func main() {
	// Scaffold the FlexBox `Main` and layout
	fMain = ui.SetupLayout()

	// Create a new application and be sure to set the root object
	app := tview.NewApplication()
	app.SetRoot(fMain, true).EnableMouse(true)

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
	go ui.UpdateCPUTemp(app, false, gtm.Cfg.UpdateInterval)
	go ui.UpdateDisk(app, true, gtm.Cfg.UpdateInterval)
	if gtm.HasGPU() && runtime.GOOS != "darwin" {
		slog.Info("Dedicated GPU enabled; setting up GPU & GPUTemp UI " +
			"goroutines ...")
		go ui.UpdateGPU(app, true, gtm.Cfg.UpdateInterval)
		go ui.UpdateGPUTemp(app, true, gtm.Cfg.UpdateInterval)
	}
	go ui.UpdateMemory(app, true, gtm.Cfg.UpdateInterval)
	go ui.UpdateNetwork(app, true, gtm.Cfg.UpdateInterval)
	go ui.UpdateProcesses(app, true, gtm.Cfg.UpdateInterval)

	// TODO: REMOVE ME - this is for testing
	//delay := time.Duration(5)
	//fmt.Printf("[main.go] Starting app in %d seconds...\n", delay)
	//time.Sleep(delay * time.Second)

	// START APP
	slog.Info("Starting the app ...")
	if err := app.Run(); err != nil {
		slog.Error("Failed to run the app! " + err.Error())
		panic(err)
	}
}
