package ui

import (
	"github.com/rivo/tview"
	"gtm"
	"log/slog"
	"strconv"
	"time"
)

func UpdateMemory(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		//isResized     bool
	)

	Layout.Memory.SetBorder(showBorder).SetTitle(LblMemory)
	slog.Info("Starting `UpdateMemory()` UI goroutine ...")

	for {
		timestamp := time.Now().UnixMilli()

		memInfo := gtm.GetMemoryInfo()

		//boxSize := "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		memUsed := gtm.ConvertBytesToGB(memInfo.Used, false)
		memTotal := gtm.ConvertBytesToGB(memInfo.Total, false)
		memUsedText := strconv.FormatFloat(memUsed, 'f', 1, 64) + " GB"
		memTotalText := strconv.FormatFloat(memTotal, 'f', 1, 64) + " GB"

		timeDelta := time.Now().UnixMilli() - timestamp
		if timeDelta < update.Milliseconds() {
			time.Sleep(time.Duration(update.Milliseconds() - timeDelta))
		}
		//time.Sleep(update)

		width, height, _ = GetInnerBoxSize(Layout.Memory.Box, width, height)
		barMemoryStatsRow := memUsedText +
			InsertCenterSpacing(memUsedText, memTotalText, width, " ") +
			memTotalText

		labelRow := "Used" +
			InsertCenterSpacing("Used", "Total", width, " ") +
			"Total"
		// Get the ratio of memory used and total memory. Use the ratio to build a text bar
		memUsedRatio := memUsed / memTotal
		barTextRow := BuildHorizontalTextBar(width, memUsedRatio)

		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: do draw
			boxText = labelRow + "\n" + barTextRow + "\n" + barMemoryStatsRow
			Layout.Memory.SetText(boxText)

			if isResized {
			}
		})
	}
}
