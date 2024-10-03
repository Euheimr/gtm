package ui

import (
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v4/mem"
	"log/slog"
	"strconv"
	"time"
)

func UpdateMemory(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		isResized     bool
		width, height int
	)

	Layout.Memory.SetBorder(showBorder).SetTitle(LblMemory)
	slog.Info("Starting `UpdateMemory()` UI goroutine ...")

	for {
		width, height, isResized = GetInnerBoxSize(Layout.Memory.Box, width, height)

		memInfo, err := mem.VirtualMemory()
		if err != nil {
			slog.Error(err.Error())
		}

		//boxSize := "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		memUsed := ConvertBytesToGB(memInfo.Used, true)
		memTotal := ConvertBytesToGB(memInfo.Total, true)
		memUsedText := strconv.Itoa(int(memUsed)) + " GB"
		memTotalText := strconv.Itoa(int(memTotal)) + " GB"

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
