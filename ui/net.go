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
		boxText string
		w, h    int
		//isResized bool
	)

	Layout.Network.SetBorder(showBorder).SetTitle(LblNetwork)
	slog.Info("Starting `UpdateNetwork()` UI goroutine ...")

	for {

		time.Sleep(update)
		w, h, _ = GetInnerBoxSize(Layout.Network.Box, w, h)
		app.QueueUpdateDraw(func() {
			// TODO: do draw
			boxText = gtm.GetHostname() + "\n" +
				"col: " + strconv.Itoa(w) + ", row: " + strconv.Itoa(h)
			Layout.Network.SetText(boxText)
		})
	}
}
