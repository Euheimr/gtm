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
		isResized     bool
	)

	Layout.Network.SetDynamicColors(true)
	Layout.Network.SetBorder(showBorder).SetTitle(LblNetwork)
	slog.Info("Starting `UpdateNetwork()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, _ = GetInnerBoxSize(Layout.Network.Box, width, height)

		netInfo := gtm.GetNetworkInfo()

		boxText = gtm.GetHostname() + "\n"
		//boxText += "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		for _, iface := range netInfo {
			boxText += BuildBoxTitleRow(
				"DOWN: ", strconv.FormatUint(iface.BytesSent, 10), width, " ")
			boxText += BuildBoxTitleRow(
				"UP: ", strconv.FormatUint(iface.BytesRecv, 10), width, " ")
		}

		SleepWithTimestampDelta(timestamp, update, isResized)

		app.QueueUpdateDraw(func() {
			Layout.Network.SetText(boxText)
		})
	}
}
