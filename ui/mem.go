package ui

import (
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v4/mem"
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
		lastDataFetch time.Time
		memInfo       *mem.VirtualMemoryStat
	)

	Layout.Memory.SetDynamicColors(true)
	Layout.Memory.SetBorder(showBorder).SetTitle(LblMemory)
	slog.Info("Starting `UpdateMemory()` UI goroutine ...")

	for {
		timestamp := time.Now().UnixMilli()

		// Limit getting device data to just once a second, and NOT with every UI update
		if time.Since(lastDataFetch) >= time.Second || len(memInfo.String()) < 1 {
			memInfo = gtm.GetMemoryInfo()
			lastDataFetch = time.Now()
		}
		/// END DATA FETCH

		//boxSize := "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		memUsed := gtm.ConvertBytesToGiB(memInfo.Used, false)
		memTotal := gtm.ConvertBytesToGiB(memInfo.Total, false)
		memUsedText := strconv.FormatFloat(memUsed, 'f', 1, 64) + " GB"
		memTotalText := strconv.FormatFloat(memTotal, 'f', 1, 64) + " GB"

		width, height, _ = GetInnerBoxSize(Layout.Memory.Box, width, height)

		memoryUsedTitleRow := BuildBoxTitleRow("Used", "Total", width, " ")
		memoryStatsRow := BuildBoxTitleRow(memUsedText, memTotalText, width, " ")

		// Get the ratio of memory used and total memory. Use the ratio to build a text bar
		//memUsedRatio := memUsed / memTotal
		progressBar := BuildProgressBar(memInfo.UsedPercent/100, width, GREEN, WHITE)

		boxText = memoryUsedTitleRow + progressBar + memoryStatsRow

		timeDelta := time.Now().UnixMilli() - timestamp
		if timeDelta < update.Milliseconds() {
			time.Sleep(time.Duration(update.Milliseconds() - timeDelta))
		}

		app.QueueUpdateDraw(func() {
			// TODO: do draw
			Layout.Memory.SetText(boxText)

		})
	}
}
