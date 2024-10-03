package ui

import (
	"github.com/rivo/tview"
	"log/slog"
	"time"
)

func UpdateProcesses(app *tview.Application, showBorder bool, update time.Duration) {
	Layout.Processes.SetBorder(showBorder).SetTitle(LblProc)
	slog.Info("Starting `UpdateProcesses()` UI goroutine ...")

	for {
		// TODO: Get process info here then pass it into the app.QueueUpdateDraw()
		// 	before sleeping

		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: do draw

		})
	}
}
