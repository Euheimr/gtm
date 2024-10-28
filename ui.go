package gtm

import (
	"github.com/rivo/tview"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"
)

type cpuBox struct {
	Stats *tview.TextView
	Temp  *tview.TextView
}

type layoutMain struct {
	CPU       *cpuBox
	Disk      *tview.TextView
	GPU       *tview.TextView
	GPUTemp   *tview.TextView
	Memory    *tview.TextView
	Network   *tview.TextView
	Processes *tview.Table
}

// These constants are text formatting tags used by the tcell package
const (
	BLACK  string = "[black]"
	BLUE          = "[blue]"
	GREEN         = "[green]"
	GRAY          = "[gray]"
	RED           = "[red]"
	WHITE         = "[white]"
	YELLOW        = "[yellow]"
)

const (
	LblCPUTemp = " CPU Temp "
	LblDisk    = " HDD / SSD "
	LblGPUTemp = " GPU Temp "
	LblMemory  = " Memory "
	LblNetwork = " Network "
	LblProc    = " Processes "
)

var (
	Layout      *layoutMain
	barSymbols  = [8]string{" ", "░", "▒", "▓", "█", "[", "|", "]"}
	treeSymbols = [4]string{"│", "├", "─", "└"}
)

func init() {
	// Initialize the main Layout ASAP
	Layout = &layoutMain{
		CPU: &cpuBox{
			Stats: tview.NewTextView(),
			Temp:  tview.NewTextView(),
		},
		Disk:      tview.NewTextView(),
		GPU:       tview.NewTextView(),
		GPUTemp:   tview.NewTextView(),
		Memory:    tview.NewTextView(),
		Network:   tview.NewTextView(),
		Processes: tview.NewTable(),
	}
}

func SetupLayout() (fMain *tview.Flex) {
	slog.Info("Setting up layout ...")

	// This is the BASE box containing ALL OTHER boxes
	fMain = tview.NewFlex()
	// Ensure the base "Main" layout view is always Rows and not Columns
	fMain.SetDirection(tview.FlexRow)

	// SETUP PRIMARY LAYOUT
	/// Row 1
	flexRow1 := tview.NewFlex()

	// ROW 1 COLUMN 1
	cpuParentBox := tview.NewFlex()
	cpuParentBox.SetBorder(true).SetTitle(" " + GetCPUModel(true) + " ")
	flexRow1.AddItem(cpuParentBox.
		AddItem(Layout.CPU.Stats, 0, 5, false).
		AddItem(Layout.CPU.Temp, 0, 2, false),
		0, 6, false)

	// ROW 1 COLUMN 2
	flexRow1.AddItem(Layout.Memory, 0, 2, false)
	fMain.AddItem(flexRow1, 0, 22, false)

	/// Row 2
	flexRow2 := tview.NewFlex()
	// ROW 2 COLUMN 1
	flexRow2.AddItem(Layout.Processes, 0, 2, false)
	// FIXME: There's a weird bug here where selecting the Processes table also
	// 	selects this row too?
	// ROW 2 COLUMN 2
	flexRow2.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(Layout.Network, 0, 2, false).
		AddItem(Layout.Disk, 0, 2, false),
		0, 1, false)
	if HasGPU() {
		// ROW 2 COLUMN 3
		flexRow2.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(Layout.GPU, 0, 4, false).
			AddItem(Layout.GPUTemp, 0, 4, false),
			0, 1, false)
	}
	fMain.AddItem(flexRow2, 0, 40, false)

	/// Row 3
	flexRow3 := tview.NewFlex()
	flexRow3.AddItem(tview.NewTextView().
		SetText(" <F1> Test   <F2> Test 1   <F3> Test 2   <F4> Test 3"),
		0, 1, false)
	fMain.AddItem(flexRow3, 0, 1, false)

	return fMain
}

func sleepWithTimestampDelta(timestamp time.Time, update time.Duration, isResized bool) {
	if isResized {
		// When the window/box primitive is resized, refresh the window info ASAP
		//slog.Debug("sleep SKIP")
		time.Sleep(0)
	} else {
		// Only sleep window refresh/updates when the window is NOT resized.
		timeDelta := time.Now().UnixMilli() - timestamp.UnixMilli()
		if timeDelta == 0 {
			//slog.Debug("sleep update = " + strconv.Itoa(int(update.Milliseconds())))
			time.Sleep(update)
		} else if timeDelta < update.Milliseconds() {
			//slog.Debug("sleep timeDelta = " + strconv.Itoa(int(update.Milliseconds()-timeDelta)))
			time.Sleep(time.Duration(update.Milliseconds() - timeDelta))
		}
	}
}

func buildProgressBar(ratio float64, columns int, colorFill string, colorEmpty string) string {
	var (
		countFill  int    = 0
		countEmpty        = columns   // default char count to total box columns
		barText    string = colorFill // insert "used" color tag here
		charUsed          = barSymbols[4]
		charEmpty         = barSymbols[1]
		charStart         = barSymbols[4]
		charEnd           = barSymbols[1]
	)
	// We never want a ratio higher than 1.0 or the bar will overflow to the next line
	if ratio <= 1.0 {
		countFill = int(math.Round(float64(columns) * ratio))
	} else {
		countFill = int(math.Round(float64(columns) * 1.0)) // Clamp the ratio to 1.0
	}

	if countFill >= 1 {
		barText += charStart
		for i := range countFill {
			if i != (countFill - 1) {
				// If we aren't on the last element, build a bar of "used" memory
				barText += charUsed
			}
		}
		countEmpty -= countFill
	}
	// Add in the second color tag for the empty or "unused" portion of the bar
	barText += colorEmpty

	// Iterate over an integer count of empty chars to add in the empty/unused part of
	//	the bar
	for i := 0; i < (countEmpty - 1); i++ {
		barText += charEmpty
	}
	return barText + charEnd + WHITE + "\n" // Cap off the end of the bar and return
}

func getInnerBoxSize(box *tview.Box, oldWidth int, oldHeight int) (width int, height int,
	isResized bool) {

	_, _, width, height = box.GetInnerRect()
	isResized = false

	if (oldWidth != 0 || oldHeight != 0) && (oldWidth != width || oldHeight != height) {
		isResized = true
		boxTitle := strings.TrimSpace(box.GetTitle())

		slog.Debug(boxTitle + " inner box size changed from (" +
			strconv.Itoa(oldWidth) + "->" + strconv.Itoa(width) + ") columns " +
			"and (" + strconv.Itoa(oldHeight) + "->" + strconv.Itoa(height) + ") rows !")
	}
	return width, height, isResized
}

func insertCenterSpacing(arg1 string, arg2 string, boxWidth int,
	spaceChar string) (spaces string) {

	spacingCount := boxWidth - len(arg1) - len(arg2)
	for range spacingCount {
		spaces += spaceChar
	}
	return spaces
}

func buildBoxTitleRow(title string, statStr string, boxWidth int, spaceChar string) string {
	return title + insertCenterSpacing(title, statStr, boxWidth, spaceChar) + statStr + "\n"
}

//////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////// UI GOROUTINES START HERE ///////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////

func UpdateCPU(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)
	Layout.CPU.Stats.SetDynamicColors(true)
	Layout.CPU.Stats.SetBorder(showBorder)
	slog.Info("Starting `UpdateCPU()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(Layout.CPU.Stats.Box, width, height)

		// TODO: use 2 boxes as columns (side-by-side) to display a graph and stats
		// 	(in that order)
		//boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height) + "\n"

		sleepWithTimestampDelta(timestamp, update, isResized)

		app.QueueUpdateDraw(func() {
			Layout.CPU.Stats.SetText(boxText)
		})
	}
}

func UpdateCPUTemp(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)

	Layout.CPU.Temp.SetDynamicColors(true)
	Layout.CPU.Temp.SetBorder(showBorder).SetTitle(LblCPUTemp)
	slog.Info("Starting `UpdateCPUTemp()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, _ = getInnerBoxSize(Layout.CPU.Temp.Box, width, height)

		//boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height) + "\n"

		sleepWithTimestampDelta(timestamp, update, isResized)

		app.QueueUpdateDraw(func() {
			Layout.CPU.Temp.SetText(boxText)
		})
	}
}

func UpdateDisk(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
		//disksVirtualStr []bool
	)

	Layout.Disk.SetDynamicColors(true)
	Layout.Disk.SetBorder(showBorder).SetTitle(LblDisk)
	slog.Info("Starting `UpdateDisk()` UI goroutine ...")

	for {
		timestamp := time.Now()

		diskInfo = GetDiskInfo()
		boxText = ""

		for _, dsk := range diskInfo {
			var diskCapacityStr string
			diskCapacity := ConvertBytesToGiB(dsk.Total, false)
			if diskCapacity < 999 {
				diskCapacityStr = strconv.FormatFloat(diskCapacity, 'f', 1, 64) + "GB"
			} else {
				diskCapacityStr = strconv.FormatFloat(diskCapacity/100.0, 'f', 1, 64) + "TB"
			}

			//boxText += dsk.Mountpoint + " | " + strconv.FormatBool(dsk.IsVirtualDisk) +
			//	" | " + strconv.FormatFloat(dsk.UsedPercent, 'g', -1, 64) +
			//	"% of " + diskCapacityStr + "\n"
			width, height, _ = getInnerBoxSize(Layout.Disk.Box, width, height)
			boxText += buildBoxTitleRow(dsk.Mountpoint, diskCapacityStr, width, " ")
			boxText += buildProgressBar(dsk.UsedPercent, width, RED, WHITE)
			boxText += "width=" + strconv.Itoa(width) + ", height=" + strconv.Itoa(height) + "\n"
		}

		sleepWithTimestampDelta(timestamp, update, isResized)

		app.QueueUpdateDraw(func() {
			Layout.Disk.SetText(boxText)
		})
	}
}

func UpdateGPU(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)

	Layout.GPU.SetDynamicColors(true)
	Layout.GPU.SetBorder(showBorder).SetTitle(" " + GetGPUName() + " ")
	slog.Info("Starting `UpdateGPU()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, _ = getInnerBoxSize(Layout.GPU.Box, width, height)

		gpuData = GetGPUInfo()
		lastElement := len(gpuData) - 1
		/// END DATA FETCH

		gpuLoadStr := strconv.FormatInt(int64(gpuData[lastElement].Load*100.0), 10) + "%"
		gpuLoadTitleRow := buildBoxTitleRow("Load:", gpuLoadStr, width, " ")

		gpuMemoryUsageRatio := gpuData[lastElement].MemoryUsage / gpuData[lastElement].MemoryTotal
		gpuMemoryStr := strconv.FormatInt(int64(gpuMemoryUsageRatio*100), 10) + "%"
		gpuMemoryTitleRow := buildBoxTitleRow("Mem:", gpuMemoryStr, width, " ")

		boxText = gpuLoadTitleRow + buildProgressBar(gpuData[lastElement].Load, width, GREEN, WHITE)
		boxText += "\n" // add an extra line gap to visually and obviously separate the info
		boxText += gpuMemoryTitleRow + buildProgressBar(gpuMemoryUsageRatio, width, GREEN, WHITE)

		sleepWithTimestampDelta(timestamp, update, isResized)

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
		width, height, _ = getInnerBoxSize(Layout.GPUTemp.Box, width, height)

		gpuData = GetGPUInfo()
		lastElement := len(gpuData) - 1

		//boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		gpuTempStr := strconv.Itoa(int(gpuData[lastElement].Temperature)) + "°C"
		gpuTempTitle := buildBoxTitleRow("Temp:", gpuTempStr, width, " ")

		boxText = gpuTempTitle + buildProgressBar(
			float64(gpuData[lastElement].Temperature)/100.0, width, GREEN, WHITE)

		sleepWithTimestampDelta(timestamp, update, isResized)

		app.QueueUpdateDraw(func() {
			Layout.GPUTemp.SetText(boxText)
		})
	}
}

func UpdateMemory(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)

	Layout.Memory.SetDynamicColors(true)
	Layout.Memory.SetBorder(showBorder).SetTitle(LblMemory)
	slog.Info("Starting `UpdateMemory()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(Layout.Memory.Box, width, height)

		memInfo = GetMemoryInfo()
		/// END DATA FETCH

		memUsed := ConvertBytesToGiB(memInfo.Used, false)
		memUsedText := strconv.FormatFloat(memUsed, 'f', 1, 64) + " GB"

		memTotal := ConvertBytesToGiB(memInfo.Total, false)
		memTotalText := strconv.FormatFloat(memTotal, 'f', 1, 64) + " GB"

		memoryUsedTitleRow := buildBoxTitleRow("Used", "Total", width, " ")
		progressBar := buildProgressBar(memInfo.UsedPercent/100, width, GREEN, WHITE)
		memoryStatsRow := buildBoxTitleRow(memUsedText, memTotalText, width, " ")

		boxText = memoryUsedTitleRow + progressBar + memoryStatsRow

		sleepWithTimestampDelta(timestamp, update, isResized)

		app.QueueUpdateDraw(func() {
			// TODO: do draw
			Layout.Memory.SetText(boxText)

		})
	}
}

func UpdateNetwork(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)

	Layout.Network.SetDynamicColors(true)
	Layout.Network.SetBorder(showBorder).SetTitle(LblNetwork)
	slog.Info("Starting `UpdateNetwork()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, _ = getInnerBoxSize(Layout.Network.Box, width, height)

		netInfo = GetNetworkInfo()

		boxText = GetHostname() + "\n"
		//boxText += "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		for _, iface := range netInfo {
			boxText += buildBoxTitleRow(
				"DOWN: ", strconv.FormatUint(iface.BytesSent, 10), width, " ")
			boxText += buildBoxTitleRow(
				"UP: ", strconv.FormatUint(iface.BytesRecv, 10), width, " ")
		}

		sleepWithTimestampDelta(timestamp, update, isResized)

		app.QueueUpdateDraw(func() {
			Layout.Network.SetText(boxText)
		})
	}
}

func UpdateProcesses(app *tview.Application, showBorder bool, update time.Duration) {

	Layout.Processes.SetBorder(showBorder).SetTitle(LblProc)
	slog.Info("Starting `UpdateProcesses()` UI goroutine ...")

	for {
		// TODO: Get process info here then pass it into the app.QueueUpdateDraw()
		// 	before sleeping

		time.Sleep(update)
		app.QueueUpdateDraw(func() {

		})
	}
}
