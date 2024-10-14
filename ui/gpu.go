package ui

import (
	"github.com/rivo/tview"
	"gtm"
	"log/slog"
	"strconv"
	"time"
)

func UpdateGPU(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		//isResized bool
	)

	gpuName := gtm.GetGPUName()

	//slog.Debug("length of gpu.GraphicsCards[] is: " + gpu.String())
	Layout.GPU.SetBorder(showBorder).SetTitle(" " + gpuName + " ")
	slog.Info("Starting `UpdateGPU()` UI goroutine ...")

	for {
		timestamp := time.Now().UnixMilli()

		time.Sleep(update)
		width, height, _ = GetInnerBoxSize(Layout.GPU.Box, width, height)

		app.QueueUpdateDraw(func() {
			// TODO: do draw
			// boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height) +
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
