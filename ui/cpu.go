package ui

import (
	"github.com/rivo/tview"
	"log/slog"
	"strconv"
	"time"
)

func UpdateCPU(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText string
		w, h    int
		//isResized bool
	)

	// TODO: get CPU name and speed then update label to that?
	Layout.CPU.Stats.SetBorder(showBorder).SetTitle(LblCPU)
	slog.Info("Starting `UpdateCPU()` UI goroutine ...")

	for {
		w, h, _ = GetInnerBoxSize(Layout.CPU.Stats.Box, w, h)

		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: use 2 boxes as columns (side-by-side) to display a graph and stats
			// 	(in that order)
			boxText = "col: " + strconv.Itoa(w) + ", row: " + strconv.Itoa(h) + "\n"
			Layout.CPU.Stats.SetText(boxText)
		})
	}
}

func UpdateCPUTemp(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText string
		w, h    int
		//isResized bool
	)

	Layout.CPU.Temp.SetBorder(showBorder).SetTitle(LblCPUTemp)
	slog.Info("Starting `UpdateCPUTemp()` UI goroutine ...")

	for {
		w, h, _ = GetInnerBoxSize(Layout.CPU.Stats.Box, w, h)

		time.Sleep(update)
		app.QueueUpdateDraw(func() {

			boxText = "col: " + strconv.Itoa(w) + ", row: " + strconv.Itoa(h) + "\n"
			Layout.CPU.Temp.SetText(boxText)

		})
	}
}
