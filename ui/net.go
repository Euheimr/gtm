package ui

import (
	"github.com/rivo/tview"
	"gtm"
	"log/slog"
	"strconv"
	"time"
)

func UpdateNetwork(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		//isResized bool
	)
	Layout.Network.SetDynamicColors(true)
	Layout.Network.SetBorder(showBorder).SetTitle(LblNetwork)
	slog.Info("Starting `UpdateNetwork()` UI goroutine ...")

	for {
		timestamp := time.Now()

		time.Sleep(update)
		width, height, _ = GetInnerBoxSize(Layout.Network.Box, width, height)

		boxText = gtm.GetHostname() + "\n" +
			"col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)

		SleepWithTimestampDelta(timestamp, update)

		app.QueueUpdateDraw(func() {
			Layout.Network.SetText(boxText)
		})
	}
}
