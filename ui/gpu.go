package ui

import (
	"github.com/rivo/tview"
	"gtm"
	"log/slog"
	"strconv"
	"time"
)

var gpuData []gtm.GPUData

func UpdateGPU(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		//isResized bool
		lastDataFetch time.Time
	)

	gpuName := gtm.GetGPUName()

	//slog.Debug("length of gpu.GraphicsCards[] is: " + gpu.String())
	Layout.GPU.SetDynamicColors(true)
	Layout.GPU.SetBorder(showBorder).SetTitle(" " + gpuName + " ")
	slog.Info("Starting `UpdateGPU()` UI goroutine ...")

	for {
		timestamp := time.Now().UnixMilli()
		width, height, _ = GetInnerBoxSize(Layout.GPU.Box, width, height)

		// Limit getting device data to just once a second, and NOT with every UI update
		if time.Since(lastDataFetch) >= time.Second || len(gpuData) < 1 {
			gpuData = gtm.GetGPUInfo()
			lastDataFetch = time.Now()
		}
		lastElement := len(gpuData) - 1
		/// END DATA FETCH

		// boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height) +
		//	"\n" +  gpuData[lastElement].String()
		gpuLoadStr := strconv.FormatInt(int64(gpuData[lastElement].Load*100.0), 10) + "%"
		gpuLoadTitleRow := BuildBoxTitleRow("Load:", gpuLoadStr, width, " ")

		gpuMemoryUsageRatio := gpuData[lastElement].MemoryUsage / gpuData[lastElement].MemoryTotal
		gpuMemoryStr := strconv.FormatInt(int64(gpuMemoryUsageRatio*100), 10) + "%"
		gpuMemoryTitleRow := BuildBoxTitleRow("Mem:", gpuMemoryStr, width, " ")

		boxText = gpuLoadTitleRow + BuildProgressBar(gpuData[lastElement].Load, width, GREEN, WHITE)
		boxText += "\n" // add an extra line gap to visually and obviously separate the info
		boxText += gpuMemoryTitleRow + BuildProgressBar(gpuMemoryUsageRatio, width, GREEN, WHITE)

		timeDelta := time.Now().UnixMilli() - timestamp
		if timeDelta < update.Milliseconds() {
			time.Sleep(time.Duration(update.Milliseconds() - timeDelta))
		}

		app.QueueUpdateDraw(func() {
			Layout.GPU.SetText(boxText)
		})
	}
}

func UpdateGPUTemp(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		//isResized bool
	)

	Layout.GPUTemp.SetDynamicColors(true)
	Layout.GPUTemp.SetBorder(showBorder).SetTitle(LblGPUTemp)
	slog.Info("Starting `UpdateGPUTemp()` UI goroutine ...")

	for {
		width, height, _ = GetInnerBoxSize(Layout.GPUTemp.Box, width, height)

		time.Sleep(update)
		lastElement := len(gpuData) - 1

		//boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		gpuTempStr := strconv.Itoa(int(gpuData[lastElement].Temperature)) + "°C"
		gpuTempTitle := BuildBoxTitleRow("Temp:", gpuTempStr, width, " ")

		boxText = gpuTempTitle + BuildProgressBar(
			float64(gpuData[lastElement].Temperature)/100.0, width, GREEN, WHITE)
		app.QueueUpdateDraw(func() {
			Layout.GPUTemp.SetText(boxText)
		})
	}
}
