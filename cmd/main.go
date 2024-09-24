package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gtm/internal"
	"log/slog"
)

var fMain *tview.Flex

func main() {
	// Logging will not work as expected unless we set it first before ANYTHING
	internal.SetupLogging()

	// Scaffold the FlexBox `Main` and layout
	fMain = internal.SetupLayout()

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

	// Start the goroutines handling the drawing of each box here
	slog.Debug("Setting up UI goroutines ...")
	go internal.UpdateCPU(app, internal.Cfg.UpdateInterval)
	go internal.UpdateCPUTemp(app, internal.Cfg.UpdateInterval)
	if internal.Cfg.EnableGPU {
		slog.Debug("Dedicated GPU enabled; setting up GPU & GPUTemp UI goroutines ...")
		go internal.UpdateGPU(app, internal.Cfg.UpdateInterval)
		go internal.UpdateGPUTemp(app, internal.Cfg.UpdateInterval)
	}
	go internal.UpdateMemory(app, internal.Cfg.UpdateInterval)
	go internal.UpdateNetwork(app, internal.Cfg.UpdateInterval)
	go internal.UpdateProcesses(app, internal.Cfg.UpdateInterval)

	// TODO: REMOVE ME - this is for testing
	//delay := time.Duration(5)
	//fmt.Printf("[main.go] Starting app in %d seconds...\n", delay)
	//time.Sleep(delay * time.Second)

	// START APP
	slog.Debug("Starting the app ...")
	//if err := app.Run(); err != nil {
	//	panic(err)
	//}

}
