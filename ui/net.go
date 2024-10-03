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
		boxText   string
		isResized bool
		w, h      int
	)

	Layout.Network.SetBorder(showBorder).SetTitle(LblNetwork)
	slog.Info("Starting `UpdateNetwork()` UI goroutine ...")

	for {
		w, h, isResized = GetInnerBoxSize(Layout.Network.Box, w, h)

		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: do draw
			boxText = gtm.GetHostname() + "\n" +
				"col: " + strconv.Itoa(w) + ", row: " + strconv.Itoa(h)

			if isResized {
				Layout.Network.SetText(boxText)
			}
		})
	}
}
