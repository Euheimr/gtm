package internal

import (
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v4/host"
	"log/slog"
	"time"
)

type layoutMain struct {
	CPU       *tview.Box
	CPUTemp   *tview.Box
	Memory    *tview.TextView
	Processes *tview.Table
	Network   *tview.Box
	GPU       *tview.Box
	GPUTemp   *tview.Box
}

// these constants are text formatting tags used by the tcell package
const (
	BLACK  string = "[black]"
	BLUE          = "[blue]"
	GREEN         = "[green]"
	RED           = "[red]"
	WHITE         = "[white]"
	YELLOW        = "[yellow]"
	GRAY          = "[gray]"
)

const (
	lblCPUTemp = "[ CPU Temp ]"
	lblGPUTemp = "[ GPU Temp ]"
	lblMemory  = "[ Memory ]"
	lblNetwork = "[ Network ]"
)

var (
	layout  layoutMain
	lblCPU  = "[" + "DETECT CPU" + "]"
	lblGPU  = "[" + "DETECT GPU" + "]"
	lblProc = "[ Processes ]"
)

func init() {
	// Initialize the main layout ASAP
	layout = layoutMain{
		CPU:       tview.NewBox(),
		CPUTemp:   tview.NewBox(),
		Memory:    tview.NewTextView(),
		Processes: tview.NewTable(),
		Network:   tview.NewBox(),
		GPU:       tview.NewBox(),
		GPUTemp:   tview.NewBox(),
	}

}

func getBoxSize(box *tview.Box) (width int, height int) {
	_, _, width, height = box.GetInnerRect()
	return width, height
}

func getHostInfo() *host.InfoStat {
	info, err := host.Info()

	if err != nil {
		slog.Error("Failed to get host.Info() !")
	}
	return info
}

func SetupLayout() (fMain *tview.Flex) {
	slog.Debug("Setting up layout ...")

	// This is the BASE box containing ALL OTHER boxes
	fMain = tview.NewFlex()
	// Ensure the base "Main" layout view is always Rows and not Columns
	fMain.SetDirection(tview.FlexRow)

	// SETUP PRIMARY LAYOUT
	flexRow1 := tview.NewFlex()

	// ROW 1 COLUMN 1
	flexRow1.AddItem(layout.CPU, 0, 8, false)
	// ROW 1 COLUMN 2
	flexRow1.AddItem(layout.Memory, 0, 4, false)
	fMain.AddItem(flexRow1, 0, 22, false)

	flexRow2 := tview.NewFlex()
	// ROW 2 COLUMN 1
	flexRow2.AddItem(layout.Processes, 0, 2, false)
	// FIXME: There's a weird bug here where selecting the Processes table also
	// 	selects this row too?
	// ROW 2 COLUMN 2
	flexRow2.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(layout.CPUTemp, 0, 2, false).
		AddItem(layout.Network, 0, 2, false),
		0, 1, false)
	if Cfg.EnableGPU {
		// ROW 2 COLUMN 3
		flexRow2.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(layout.GPU, 0, 4, false).
			AddItem(layout.GPUTemp, 0, 4, false),
			0, 1, false)
	}
	fMain.AddItem(flexRow2, 0, 40, false)

	flexRow3 := tview.NewFlex()
	// ROW 3
	flexRow3.AddItem(tview.NewTextView().
		SetText(" <F1> Test   <F2> Test 1   <F3> Test 2   <F4> Test 3"),
		0, 1, false)
	fMain.AddItem(flexRow3, 0, 1, false)

	return fMain
}

func UpdateCPU(app *tview.Application, update time.Duration) {
	// TODO: get CPU name and speed then update label to that?
	layout.CPU.SetBorder(true).SetTitle(lblCPU)
	slog.Debug("Starting `UpdateCPU()` goroutine ...")

	for {
		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: use 2 boxes as columns (side-by-side) to display a graph and stats
			// 	(in that order)

		})
	}
}

func UpdateCPUTemp(app *tview.Application, update time.Duration) {
	layout.CPUTemp.SetBorder(true).SetTitle(lblCPUTemp)
	slog.Debug("Starting `UpdateCPUTemp()` goroutine ...")

	for {
		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: do draw
		})
	}
}

func UpdateGPU(app *tview.Application, update time.Duration) {
	layout.GPU.SetBorder(true).SetTitle(lblGPU)
	slog.Debug("Starting `UpdateGPU()` goroutine ...")

	for {
		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: do draw
		})
	}
}

func UpdateGPUTemp(app *tview.Application, update time.Duration) {
	layout.GPUTemp.SetBorder(true).SetTitle(lblGPUTemp)
	slog.Debug("Starting `UpdateGPUTemp()` goroutine ...")

	for {
		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: do draw
		})
	}
}

func UpdateMemory(app *tview.Application, update time.Duration) {
	layout.Memory.SetBorder(true).SetTitle(lblMemory)
	slog.Debug("Starting `UpdateMemory()` goroutine ...")

	for {
		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: do draw
		})
	}
}

func UpdateNetwork(app *tview.Application, update time.Duration) {
	layout.Network.SetBorder(true).SetTitle(lblNetwork)
	slog.Debug("Starting `UpdateNetwork()` goroutine ...")

	for {
		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: do draw
		})
	}
}

func UpdateProcesses(app *tview.Application, update time.Duration) {
	layout.Processes.SetBorder(true).SetTitle(lblProc)
	slog.Debug("Starting `UpdateProcesses()` goroutine ...")

	for {
		time.Sleep(update)
		app.QueueUpdateDraw(func() {
			// TODO: do draw

		})
	}
}
