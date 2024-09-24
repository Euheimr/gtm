package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gtm/internal"
)

var fMain *tview.Flex

func main() {
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
	go internal.UpdateCPU(app, internal.Cfg.UpdateInterval)
	go internal.UpdateCPUTemp(app, internal.Cfg.UpdateInterval)
	if internal.Cfg.EnableGPU {
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
	if err := app.Run(); err != nil {
		panic(err)
	}

}
