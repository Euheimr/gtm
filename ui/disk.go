package ui

import (
	"github.com/rivo/tview"
	"log/slog"
	"strconv"
	"time"
)

func UpdateDisk(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText string
		w, h    int
		//isResized bool
	)

	Layout.Disk.SetBorder(showBorder).SetTitle(LblDisk)
	slog.Info("Starting `UpdateDisk()` UI goroutine ...")

	for {
		w, h, _ = GetInnerBoxSize(Layout.Disk.Box, w, h)

		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: do draw
			boxText = "col: " + strconv.Itoa(w) + ", row: " + strconv.Itoa(h)
			Layout.Disk.SetText(boxText)
		})
	}
}
