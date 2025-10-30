package main

import (
	"gtm"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type CPUBox struct {
	Stats *tview.TextView
	Temp  *tview.TextView
}

type GPUBox struct {
	Stats *tview.TextView
	Temp  *tview.TextView
}

type LayoutMain struct {
	CPU       *CPUBox
	Disk      *tview.TextView
	GPU       *GPUBox
	Memory    *tview.TextView
	Network   *tview.TextView
	Processes *tview.Table
}

var (
	fMain          *tview.Flex
	layout         *LayoutMain
	StartTimestamp time.Time
)

func init() {
	StartTimestamp = time.Now()

	// Read the `.env` config before logging and anything else
	if err := gtm.ReadConfig(); err != nil {
		log.Fatal(err)
	}

	// Logging will not work as expected unless we set it first, but only after reading
	//	`.env` config
	gtm.SetupFileLogging()

	if gtm.Cfg.Debug {
		// This is for performance profiling (pprof). Open a web browser and see:
		//	http://localhost:6060/debug/pprof/
		go func() {
			// For docs, see: https://pkg.go.dev/runtime/pprof and:
			//	https://github.com/google/pprof/blob/main/doc/README.md
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	// Seed the initial values & data before setting up the rest of the app
	gtm.HostInfo()
	gtm.CPUInfo()
	gtm.GetCPUStats()
	gtm.DisksStats()
	gtm.MemoryStats()
	gtm.NetworkStats()
	if gtm.HasGPU {
		gtm.GPUStats()
	}

	// Initialize the main layout ASAP
	layout = &LayoutMain{
		CPU: &CPUBox{
			Stats: tview.NewTextView(),
		},
		Disk:      tview.NewTextView(),
		Memory:    tview.NewTextView(),
		Network:   tview.NewTextView(),
		Processes: tview.NewTable(),
	}
	if gtm.IsAdmin {
		layout.CPU.Temp = tview.NewTextView()
	}

	if gtm.HasGPU {
		layout.GPU = &GPUBox{
			Stats: tview.NewTextView(),
			Temp:  tview.NewTextView(),
		}
	}
}

func setupLayout() {
	slog.Info("Setting up layout ...")

	// This is the BASE box containing ALL OTHER boxes
	fMain = tview.NewFlex()
	// Ensure the base "Main" layout view is always Rows and not Columns
	fMain.SetDirection(tview.FlexRow)

	// SETUP PRIMARY LAYOUT
	// ROW 1 COLUMN 1
	cpuParentBox := tview.NewFlex()
	layout.CPU.Stats.SetTitle(" " + gtm.CPUModelName() + " ")

	cpuParentBox.AddItem(layout.CPU.Stats, 0, 5, false)
	if gtm.IsAdmin {
		cpuParentBox.AddItem(layout.CPU.Temp, 0, 2, false)
	}
	flexRow1 := tview.NewFlex()
	flexRow1.AddItem(cpuParentBox, 0, 6, false)

	// ROW 1 COLUMN 2
	flexRow1.AddItem(layout.Memory, 0, 2, false)

	/// Row 2
	flexRow2 := tview.NewFlex()
	// ROW 2 COLUMN 1
	flexRow2.AddItem(layout.Processes, 0, 2, false)
	// FIXME: There's a weird bug here where selecting the Processes table also
	// 	selects this row too?
	// ROW 2 COLUMN 2
	flexRow2.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(layout.Network, 0, 2, false).
		AddItem(layout.Disk, 0, 2, false),
		0, 1, false)
	if gtm.HasGPU {
		// ROW 2 COLUMN 3
		flexRow2.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(layout.GPU.Stats, 0, 4, false).
			AddItem(layout.GPU.Temp, 0, 4, false),
			0, 1, false)
	}

	/// Row 3
	flexRow3 := tview.NewFlex()
	flexRow3.AddItem(tview.NewTextView().
		SetText(" <F1> Test   <F2> Test 1   <F3> Test 2   <F4> Test 3"),
		0, 1, false)

	fMain.AddItem(flexRow1, 0, 22, false).
		AddItem(flexRow2, 0, 40, false).
		AddItem(flexRow3, 0, 1, false)
}

func main() {
	// Scaffold the FlexBox `Main` and layout
	setupLayout()

	// Create a new application and be sure to set the root object
	app := tview.NewApplication()
	// TODO: Enable mouse clicking when adding mouse input captures in the future
	app.SetRoot(fMain, true).EnableMouse(false)

	// Setup keybinds ...
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC, tcell.KeyEscape:
			app.Stop()
		default:
			return event
		}
		return event
	})

	// We want to have the initial loading to be instant and not lag when the app first loads
	// For example: we have an update interval 1 to 5 seconds. If we first run the app, it
	// 	will take 1-5 seconds for the data & UI elements to show up. This is undesirable.
	// To fix this, we set the update interval to 1ms while the goroutines spin up, then set
	// 	it back to default right before the app.Run() call
	defaultUpdate := gtm.Cfg.UpdateInterval
	gtm.Cfg.SetUpdateInterval(time.Millisecond)

	// Setup goroutines handling the drawing of each box here
	slog.Info("Setting up UI goroutines ...")
	go gtm.UpdateCPU(app, layout.CPU.Stats, true)
	go gtm.UpdateDisk(app, layout.Disk, true)
	go gtm.UpdateMemory(app, layout.Memory, true)
	go gtm.UpdateNetwork(app, layout.Network, true)
	go gtm.UpdateProcesses(app, layout.Processes, true)
	if gtm.IsAdmin {
		go gtm.UpdateCPUTemp(app, layout.CPU.Temp, true)
	}
	if gtm.HasGPU {
		slog.Info("GPU detected! Setting up GPU/GPUTemp UI goroutines ...")
		go gtm.UpdateGPU(app, layout.GPU.Stats, true)
		go gtm.UpdateGPUTemp(app, layout.GPU.Temp, true)
	}

	slog.Info("Waiting for goroutines to start up ...")
	time.Sleep(20 * time.Millisecond) // wait to start up all the goroutines

	// START APP
	slog.Info("Starting the app ...")
	gtm.Cfg.SetUpdateInterval(defaultUpdate)
	if err := app.Run(); err != nil {
		slog.Error("Failed to run the app! " + err.Error())
		panic(err)
	}
}
