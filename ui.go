package gtm

import (
	"context"
	"github.com/rivo/tview"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"
)

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
	barSymbols  = [8]string{" ", "░", "▒", "▓", "█", "[", "|", "]"}
	treeSymbols = [4]string{"│", "├", "─", "└"}
)

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
		countEmpty        = columns   // default char count to total box columns (box width)
		barText    string = colorFill // insert "used" / "load" color tag here
		charUsed          = barSymbols[4]
		charEmpty         = barSymbols[1]
		charStart         = barSymbols[4]
		charEnd           = barSymbols[1]
	)
	// Never have a ratio higher than 1.0 (100%) or the bar will overflow to the next line
	if ratio <= 1.0 {
		countFill = int(math.Round(float64(columns) * ratio))
	} else {
		// Clamp the ratio to 1.0 ONLY if above 1.0
		countFill = int(math.Round(float64(columns) * 1.0))
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
	} else if ratio > 0.0 {
		// If we are above 0% load/usage, then always show at least 1 character indicating
		//	"used" or a "load"
		barText += charUsed
		// Also, we need to make sure the countEmpty is -1 to not overflow text to the next
		//	line
		countEmpty -= 1
	}
	// Add in the second color tag for the "empty" or "unused" portion of the bar
	barText += colorEmpty

	// Iterate over an integer count of empty chars to add in the empty/unused part of
	//	the bar. We -1 to countEmpty to make room for the last character (var charEnd)
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

////##################################################################################////
////########################//// UI GOROUTINES START HERE ////########################////

// UpdateCPU is a text UI function that starts as a goroutine before the application
// starts.
func UpdateCPU(app *tview.Application, box *tview.TextView, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)
	box.SetDynamicColors(true)
	box.SetBorder(showBorder)
	slog.Info("Starting `UpdateCPU()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		// TODO: use 2 boxes as columns (side-by-side) to display a graph and stats
		// 	(in that order)
		//boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height) + "\n"

		sleepWithTimestampDelta(timestamp, update, isResized)

		if isResized {
			// Re-draw immediately if the window is resized
			app.QueueUpdateDraw(func() {
				box.SetText(boxText)
			})
		} else {
			app.QueueUpdate(func() {
				// TODO: do draw
				box.SetText(boxText)
			})
		}
		slog.Log(context.Background(), LEVEL_PERF,
			"UpdateCPU() time: "+(time.Since(timestamp)-update).String())
	}
}

func UpdateCPUTemp(app *tview.Application, box *tview.TextView, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(LblCPUTemp)
	slog.Info("Starting `UpdateCPUTemp()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		//boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height) + "\n"

		sleepWithTimestampDelta(timestamp, update, isResized)

		if isResized {
			// Re-draw immediately if the window is resized
			app.QueueUpdateDraw(func() {
				box.SetText(boxText)
			})
		} else {
			app.QueueUpdate(func() {
				// TODO: do draw
				box.SetText(boxText)
			})
		}
		slog.Log(context.Background(), LEVEL_PERF,
			"UpdateCPUTemp() time: "+(time.Since(timestamp)-update).String())
	}
}

//// Disk/HDD/SSD ////####################################################################

func UpdateDisk(app *tview.Application, box *tview.TextView, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
		//disksVirtualStr []bool
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(LblDisk)
	slog.Info("Starting `UpdateDisk()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		disksStats = GetDisksStats()
		boxText = ""

		for _, dsk := range disksStats {
			var diskCapacityStr string
			diskCapacity := ConvertBytesToGiB(dsk.Total, false)
			if diskCapacity < 999 {
				diskCapacityStr = strconv.FormatFloat(
					diskCapacity, 'f', 1, 64) + " GB"
			} else {
				diskCapacityStr = strconv.FormatFloat(
					diskCapacity/1000.0, 'f', 2, 64) + " TB"
			}

			//boxText += dsk.Mountpoint + " | " + strconv.FormatBool(dsk.IsVirtualDisk) +
			//	" | " + strconv.FormatFloat(dsk.UsedPercent, 'g', -1, 64) +
			//	"% of " + diskCapacityStr + "\n"
			boxText += buildBoxTitleRow(dsk.Mountpoint, diskCapacityStr, width, " ")
			boxText += buildProgressBar(dsk.UsedPercent, width, RED, WHITE)
			//boxText += "width=" + strconv.Itoa(width) + ", height=" + strconv.Itoa(height) + "\n"
		}

		sleepWithTimestampDelta(timestamp, update, isResized)

		if isResized {
			// Re-draw immediately if the window is resized
			app.QueueUpdateDraw(func() {
				box.SetText(boxText)
			})
		} else {
			app.QueueUpdate(func() {
				// TODO: do draw
				box.SetText(boxText)
			})
		}
		slog.Log(context.Background(), LEVEL_PERF,
			"UpdateDisk() time: "+(time.Since(timestamp)-update).String())
	}
}

//// GPU ////#############################################################################

func UpdateGPU(app *tview.Application, box *tview.TextView, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(" " + GPUName() + " ")
	slog.Info("Starting `UpdateGPU()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		gpuStats = GetGPUStats()
		lastElement := len(gpuStats) - 1
		/// END DATA FETCH

		gpuLoadStr := strconv.FormatInt(int64(gpuStats[lastElement].Load*100.0), 10) + "%"
		gpuLoadTitleRow := buildBoxTitleRow("Load:", gpuLoadStr, width, " ")

		gpuMemoryUsageRatio := gpuStats[lastElement].MemoryUsage / gpuStats[lastElement].MemoryTotal
		gpuMemoryStr := strconv.FormatInt(int64(gpuMemoryUsageRatio*100), 10) + "%"
		gpuMemoryTitleRow := buildBoxTitleRow("Mem:", gpuMemoryStr, width, " ")

		boxText = gpuLoadTitleRow + buildProgressBar(gpuStats[lastElement].Load, width, GREEN, WHITE)
		boxText += "\n" // add an extra line gap to visually and obviously separate the info
		boxText += gpuMemoryTitleRow + buildProgressBar(gpuMemoryUsageRatio, width, GREEN, WHITE)

		sleepWithTimestampDelta(timestamp, update, isResized)

		if isResized {
			// Re-draw immediately if the window is resized
			app.QueueUpdateDraw(func() {
				box.SetText(boxText)
			})
		} else {
			app.QueueUpdate(func() {
				// TODO: do draw
				box.SetText(boxText)
			})
		}
		slog.Log(context.Background(), LEVEL_PERF,
			"UpdateGPU() time: "+(time.Since(timestamp)-update).String())
	}
}

func UpdateGPUTemp(app *tview.Application, box *tview.TextView, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(LblGPUTemp)
	slog.Info("Starting `UpdateGPUTemp()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		gpuStats = GetGPUStats()
		lastElement := len(gpuStats) - 1

		//boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		gpuTempStr := strconv.Itoa(int(gpuStats[lastElement].Temperature)) + "°C"
		gpuTempTitle := buildBoxTitleRow("Temp:", gpuTempStr, width, " ")

		boxText = gpuTempTitle + buildProgressBar(
			float64(gpuStats[lastElement].Temperature)/100.0, width, GREEN, WHITE)

		sleepWithTimestampDelta(timestamp, update, isResized)

		if isResized {
			// Re-draw immediately if the window is resized
			app.QueueUpdateDraw(func() {
				box.SetText(boxText)
			})
		} else {
			app.QueueUpdate(func() {
				// TODO: do draw
				box.SetText(boxText)
			})
		}
		slog.Log(context.Background(), LEVEL_PERF,
			"UpdateGPUTemp() time: "+(time.Since(timestamp)-update).String())
	}
}

//// Memory ////##########################################################################

func UpdateMemory(app *tview.Application, box *tview.TextView, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(LblMemory)
	slog.Info("Starting `UpdateMemory()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		memInfo = GetMemoryStats()
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

		if isResized {
			// Re-draw immediately if the window is resized
			app.QueueUpdateDraw(func() {
				box.SetText(boxText)
			})
		} else {
			app.QueueUpdate(func() {
				// TODO: do draw
				box.SetText(boxText)
			})
		}
		slog.Log(context.Background(), LEVEL_PERF,
			"UpdateMemory() time: "+(time.Since(timestamp)-update).String())
	}
}

//// Network ////#########################################################################

func UpdateNetwork(app *tview.Application, box *tview.TextView, showBorder bool, update time.Duration) {
	var (
		boxText       string
		width, height int
		isResized     bool
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(LblNetwork)
	slog.Info("Starting `UpdateNetwork()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

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

		if isResized {
			// Re-draw immediately if the window is resized
			app.QueueUpdateDraw(func() {
				box.SetText(boxText)
			})
		} else {
			app.QueueUpdate(func() {
				// TODO: do draw
				box.SetText(boxText)
			})
		}
		slog.Log(context.Background(), LEVEL_PERF,
			"UpdateNetwork() time: "+(time.Since(timestamp)-update).String())
	}
}

//// Processes ////#######################################################################

func UpdateProcesses(app *tview.Application, box *tview.Table, showBorder bool, update time.Duration) {

	box.SetBorder(showBorder).SetTitle(LblProc)
	slog.Info("Starting `UpdateProcesses()` UI goroutine ...")

	for {
		//timestamp := time.Now()
		// TODO: Get process info here then pass it into the app.QueueUpdateDraw()
		// 	before sleeping

		time.Sleep(update)
		app.QueueUpdate(func() {

		})
	}
}
