package ui

import (
	"github.com/rivo/tview"
	"log/slog"
	"time"
)

func UpdateCPU(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		//isResized bool
	)
	Layout.CPU.Stats.SetDynamicColors(true)
	Layout.CPU.Stats.SetBorder(showBorder)
	slog.Info("Starting `UpdateCPU()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, _ = GetInnerBoxSize(Layout.CPU.Stats.Box, width, height)

		// TODO: use 2 boxes as columns (side-by-side) to display a graph and stats
		// 	(in that order)
		//boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height) + "\n"

		SleepWithTimestampDelta(timestamp, update)

		app.QueueUpdateDraw(func() {
			Layout.CPU.Stats.SetText(boxText)
		})
	}
}

func UpdateCPUTemp(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		//isResized bool
	)
	Layout.CPU.Temp.SetDynamicColors(true)
	Layout.CPU.Temp.SetBorder(showBorder).SetTitle(LblCPUTemp)
	slog.Info("Starting `UpdateCPUTemp()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, _ = GetInnerBoxSize(Layout.CPU.Temp.Box, width, height)

		//boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height) + "\n"

		SleepWithTimestampDelta(timestamp, update)

		app.QueueUpdateDraw(func() {
			Layout.CPU.Temp.SetText(boxText)
		})
	}
}
