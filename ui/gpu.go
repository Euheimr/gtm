package ui

import (
	"github.com/rivo/tview"
	"gtm"
	"log/slog"
	"strconv"
	"time"
)

var gpuData []gtm.GPUInfo

func UpdateGPU(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)

	Layout.GPU.SetDynamicColors(true)
	Layout.GPU.SetBorder(showBorder).SetTitle(" " + gtm.GetGPUName() + " ")
	slog.Info("Starting `UpdateGPU()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, _ = GetInnerBoxSize(Layout.GPU.Box, width, height)

		gpuData = gtm.GetGPUInfo()
		lastElement := len(gpuData) - 1
		/// END DATA FETCH

		gpuLoadStr := strconv.FormatInt(int64(gpuData[lastElement].Load*100.0), 10) + "%"
		gpuLoadTitleRow := BuildBoxTitleRow("Load:", gpuLoadStr, width, " ")

		gpuMemoryUsageRatio := gpuData[lastElement].MemoryUsage / gpuData[lastElement].MemoryTotal
		gpuMemoryStr := strconv.FormatInt(int64(gpuMemoryUsageRatio*100), 10) + "%"
		gpuMemoryTitleRow := BuildBoxTitleRow("Mem:", gpuMemoryStr, width, " ")

		boxText = gpuLoadTitleRow + BuildProgressBar(gpuData[lastElement].Load, width, GREEN, WHITE)
		boxText += "\n" // add an extra line gap to visually and obviously separate the info
		boxText += gpuMemoryTitleRow + BuildProgressBar(gpuMemoryUsageRatio, width, GREEN, WHITE)

		SleepWithTimestampDelta(timestamp, update, isResized)

		app.QueueUpdateDraw(func() {
			Layout.GPU.SetText(boxText)
		})
	}
}

func UpdateGPUTemp(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)

	Layout.GPUTemp.SetDynamicColors(true)
	Layout.GPUTemp.SetBorder(showBorder).SetTitle(LblGPUTemp)
	slog.Info("Starting `UpdateGPUTemp()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, _ = GetInnerBoxSize(Layout.GPUTemp.Box, width, height)

		gpuData = gtm.GetGPUInfo()
		lastElement := len(gpuData) - 1

		//boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		gpuTempStr := strconv.Itoa(int(gpuData[lastElement].Temperature)) + "°C"
		gpuTempTitle := BuildBoxTitleRow("Temp:", gpuTempStr, width, " ")

		boxText = gpuTempTitle + BuildProgressBar(
			float64(gpuData[lastElement].Temperature)/100.0, width, GREEN, WHITE)

		SleepWithTimestampDelta(timestamp, update, isResized)

		app.QueueUpdateDraw(func() {
			Layout.GPUTemp.SetText(boxText)
		})
	}
}
