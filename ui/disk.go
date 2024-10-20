package ui

import (
	"github.com/rivo/tview"
	"log/slog"
	"strconv"
	"time"
)

func UpdateDisk(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		//isResized bool
	)

	Layout.Disk.SetBorder(showBorder).SetTitle(LblDisk)
	slog.Info("Starting `UpdateDisk()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, _ = GetInnerBoxSize(Layout.Disk.Box, width, height)

		time.Sleep(update)
		SleepWithTimestampDelta(timestamp, update)

		app.QueueUpdateDraw(func() {
			// TODO: do draw
			boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
			Layout.Disk.SetText(boxText)
		})
	}
}
