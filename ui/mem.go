package ui

import (
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v4/mem"
	"gtm"
	"log/slog"
	"strconv"
	"time"
)

var memInfo *mem.VirtualMemoryStat

func UpdateMemory(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		//isResized     bool
	)

	Layout.Memory.SetDynamicColors(true)
	Layout.Memory.SetBorder(showBorder).SetTitle(LblMemory)
	slog.Info("Starting `UpdateMemory()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, _ = GetInnerBoxSize(Layout.Memory.Box, width, height)

		memInfo = gtm.GetMemoryInfo()
		/// END DATA FETCH

		memUsed := gtm.ConvertBytesToGiB(memInfo.Used, false)
		memUsedText := strconv.FormatFloat(memUsed, 'f', 1, 64) + " GB"

		memTotal := gtm.ConvertBytesToGiB(memInfo.Total, false)
		memTotalText := strconv.FormatFloat(memTotal, 'f', 1, 64) + " GB"

		memoryUsedTitleRow := BuildBoxTitleRow("Used", "Total", width, " ")
		progressBar := BuildProgressBar(memInfo.UsedPercent/100, width, GREEN, WHITE)
		memoryStatsRow := BuildBoxTitleRow(memUsedText, memTotalText, width, " ")

		boxText = memoryUsedTitleRow + progressBar + memoryStatsRow

		SleepWithTimestampDelta(timestamp, update)

		app.QueueUpdateDraw(func() {
			// TODO: do draw
			Layout.Memory.SetText(boxText)

		})
	}
}
