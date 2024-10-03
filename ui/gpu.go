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
		boxText   string
		isResized bool
		w, h      int
	)

	gpuName := gtm.GetGPUName()

	//slog.Debug("length of gpu.GraphicsCards[] is: " + gpu.String())
	Layout.GPU.SetBorder(showBorder).SetTitle(" " + gpuName + " ")
	slog.Info("Starting `UpdateGPU()` UI goroutine ...")

	for {
		w, h, isResized = GetInnerBoxSize(Layout.GPU.Box, w, h)

		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: do draw
			boxText = "col: " + strconv.Itoa(w) + ", row: " + strconv.Itoa(h)

			if isResized {
				Layout.GPU.SetText(boxText)
			}
		})
	}
}

func UpdateGPUTemp(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText   string
		isResized bool
		w, h      int
	)

	Layout.GPUTemp.SetBorder(showBorder).SetTitle(LblGPUTemp)
	slog.Info("Starting `UpdateGPUTemp()` UI goroutine ...")

	for {
		w, h, isResized = GetInnerBoxSize(Layout.GPUTemp.Box, w, h)

		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: do draw
			boxText = "col: " + strconv.Itoa(w) + ", row: " + strconv.Itoa(h)

			if isResized {
				Layout.GPUTemp.SetText(boxText)
			}
		})
	}
}
