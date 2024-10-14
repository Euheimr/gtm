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

		// Limit getting device data to just once a second, and NOT with every UI update
		if time.Since(lastDataFetch) >= time.Second || len(gpuData) < 1 {
			gpuData = gtm.GetGPUInfo()
			lastDataFetch = time.Now()
		}
		lastElement := len(gpuData) - 1
		/// END DATA FETCH

		timeDelta := time.Now().UnixMilli() - timestamp
		if timeDelta < update.Milliseconds() {
			time.Sleep(time.Duration(update.Milliseconds() - timeDelta))
		}

		width, height, _ = GetInnerBoxSize(Layout.GPU.Box, width, height)

		app.QueueUpdateDraw(func() {
			// TODO: do draw
			// boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height) +
			//	"\n" +  gpuData[lastElement].String()
			loadTitle := "Load:"
			gpuLoadStr := string(strconv.FormatInt(int64(gpuData[lastElement].Load*100.0), 10)) + "%"
			gpuLoadTitleRow := loadTitle + InsertCenterSpacing(loadTitle, gpuLoadStr, width, " ") + gpuLoadStr + "\n"

			boxText = gpuLoadTitleRow + BuildHorizontalTextBar(width, gpuData[lastElement].Load, GREEN, WHITE)
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
		app.QueueUpdateDraw(func() {
			// TODO: do draw
			boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
			Layout.GPUTemp.SetText(boxText)
		})
	}
}
