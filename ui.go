package gtm

import (
	"context"
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v4/mem"
	"log/slog"
	"math"
	"slices"
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
	barSymbols = [8]string{" ", "░", "▒", "▓", "█", "[", "|", "]"}
	// ascii codes => https://theasciicode.com.ar/
	lineSymbols = [9]string{" ", "─", "│", "┌", "└", "┐", "┘", "▏", "▕"}
	//treeSymbols      = [5]string{"│", "┤", "├", "─", "└"}
	//blockSymbols     = [4]string{"█", "▄", "■", "▀"}
	//directionSymbols = [8]string{"↑", "↓", "←", "→", "↖", "↗", "↘", "↙"}
	programStartTimestamp = time.Now()
)

const (
	Null                   = '\u0000'
	LineHorizontal         = '\u2500' // ─
	LineVertical           = '\u2502' // │
	LineVerticalHeavy      = '\u2503' // ┃
	LineDownRight          = '\u250C' // ┌
	LineDownLeft           = '\u2510' // ┐
	LineUpRight            = '\u2514' // └
	LineUpLeft             = '\u2518' // ┘
	LineVerticalRight      = '\u251C' // ├
	LineVerticalLeft       = '\u2524' // ┤
	LineHorizontalUp       = '\u2534' // ┴
	LineHorizontalDown     = '\u252C' // ┬
	LineHorizontalVertical = '\u253C' // ┼
)

var timestampHistory []time.Duration

func sleepWithTimestampDelta(timestamp time.Time) {
	// Only sleep window refresh/updates when the window is NOT resized.
	timeDelta := time.Now().UnixMilli() - timestamp.UnixMilli()
	if timeDelta == 0 {
		//slog.Debug("sleep update = " + strconv.Itoa(int(update.Milliseconds())))
		time.Sleep(Cfg.UpdateInterval)
	} else if timeDelta < Cfg.UpdateInterval.Milliseconds() {
		//slog.Debug("sleep timeDelta = " + strconv.Itoa(int(update.Milliseconds()-timeDelta)))
		time.Sleep(time.Duration(Cfg.UpdateInterval.Milliseconds() - timeDelta))
	} else if timeDelta > Cfg.UpdateInterval.Milliseconds() {
		// the timeDelta is greater than the update, don't sleep and update immediately
		time.Sleep(0)
	}
}

func buildProgressBar(ratio float64, oldRatio float64, columns int, colorFill string, colorEmpty string) string {
	var (
		countFill    int    = 0
		oldCountFill        = 0
		countEmpty          = columns // default char count to total box columns (box width)
		barText      string = ""
		charUsed            = barSymbols[4]
		charOld             = barSymbols[2]
		charEmpty           = barSymbols[1]
		charStart           = barSymbols[4]
		charEnd             = barSymbols[1]
	)

	// Color the bar based on how full they are
	if ratio >= 0.85 {
		barText = RED
	} else if ratio >= 0.7 && ratio < 0.85 {
		barText = YELLOW
	} else {
		barText = colorFill
	}

	// If the ratio is greater than 1.0, then a programming error has occurred and I need
	//	to know about it ASAP
	if ratio > 1.0 {
		panic("progress bar ratio cannot be greater than 1.0")
	}

	if ratio <= 1.0 && oldRatio <= 1.0 {
		countFill = int(math.Round(float64(columns) * ratio))
		oldCountFill = int(math.Round(float64(columns) * oldRatio))
	} else {
		// Never have a ratio higher than 1.0 (100%) or the bar will overflow to the next line
		// So, let's clamp the ratio to 1.0 ONLY if above 1.0
		countFill = int(math.Round(float64(columns) * 1.0))
		oldCountFill = int(math.Round(float64(columns) * 1.0))
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
	if oldCountFill > countFill {
		barText += RED
		positiveDelta := oldCountFill - countFill
		for range positiveDelta {
			barText += charOld
		}
		countEmpty -= positiveDelta
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

func buildBoxTitleStatRow(title string, statStr string, boxWidth int, spaceChar string) string {
	return title + insertCenterSpacing(title, statStr, boxWidth, spaceChar) + statStr + "\n"
}

func buildBoxTitleCentered(title string, color string, boxWidth int, spaceChar string) string {
	var titleString string

	spacesCount := boxWidth - len(title)
	for range spacesCount / 2 {
		titleString += spaceChar
	}
	titleString += color + title + WHITE
	for range spacesCount / 2 {
		titleString += spaceChar
	}
	return titleString + "\n"
}

func mean(timestamps []time.Duration) float64 {
	total := 0.0

	for _, t := range timestamps {
		total += float64(t.Microseconds())
	}
	return total / float64(len(timestamps))
}

func buildGraph(stat any, boxWidth int, boxHeight int) (graph string) {
	timestamp := time.Now()
	var graphData []int

	switch stat.(type) {
	case []CPUStat:
		for _, stat := range stat.([]CPUStat) {
			graphData = append(graphData, int(math.Round((stat.UsagePercent/100)*float64(boxHeight))))

			var dataStr string
			for i, dat := range graphData {
				if i == 0 {
					dataStr += strconv.Itoa(dat)
				} else {
					dataStr += "," + strconv.Itoa(dat)
				}
			}
			//slog.Debug("buildGraph(): graphData = " + dataStr)
		}
	case []CPUTempStat:

	case []DiskStat:

	case []GPUStat:
		//return "GPU graph"
	}

	if len(graphData) > 0 {
		// limit the length of the CPU stat data to the width of the box so that the line
		//	graph does not overflow
		if len(graphData) > boxWidth {
			graphData = graphData[boxWidth:]
		}

		maxValue := slices.Max(graphData)
		minValue := slices.Min(graphData)

		for r := boxHeight; r > 0; r-- {
			if r < minValue || r > maxValue {
				//for range boxWidth {
				//	graphStr += " "
				//}
				graph = "\n" + graph
			} else {
				// color the graph
				graph += GREEN
				// "│", "─", "┌", "┐", "└", "┘"
				for _, row := range graphData {
					if row == r {
						graph += lineSymbols[1]
					} else {
						graph += " "
					}
				}
				graph += "\n"
			}
		}
	} else {
		return "No graph data"
	}

	duration := time.Since(timestamp)
	if duration != 0 {
		timestampHistory = append(timestampHistory, duration)
		slog.Log(context.Background(), LevelPerf,
			"buildGraph() time: "+(duration).String()+", "+
				"mean time: "+strconv.FormatFloat(mean(timestampHistory), 'f', 2, 64)+"µs")
	}
	return graph
}

func buildGraphMatrix(stat any, boxWidth int, boxHeight int) (graph string) {
	timestamp := time.Now()

	var matrix [][]int32

	// The first index of the matrix is the height, the second index is the width and the
	//	value at [x][y] equals the index of the symbol to use for the graph string
	matrix = make([][]int32, boxHeight+1)
	for row := range matrix {
		matrix[row] = make([]int32, boxWidth)
	}

	switch stat.(type) {
	case []CPUStat:
		// make sure we just take the last n stats for the matrix so we don't overflow index
		if len(stat.([]CPUStat)) > boxWidth {
			stat = stat.([]CPUStat)[len(stat.([]CPUStat))-boxWidth:]
		}

		for key, val := range stat.([]CPUStat) {
			row := int(math.Round((val.UsagePercent / 100) * float64(boxHeight)))
			matrix[row][key] = LineHorizontal
		}
	case []CPUTempStat:
	case []DiskStat:
	case []GPUStat:
	}

	for col := 0; col < (boxWidth - 1); col++ { // left -> right
		// ?? i need to do something here?
		previousRow := 0

		for row := 0; row < boxHeight; row++ { // bottom -> up
			if previousRow == 0 {
				if matrix[row][col] == LineHorizontal && matrix[row][col+1] == 0 {
					// remember the row then search for the ceiling?
					previousRow = row
				}
			} else if matrix[row][col] == 0 && matrix[row][col+1] == 1 {
				rowDelta := previousRow - row
				for r := row; r < rowDelta; r++ {
					matrix[r][col] = 2
				}
			}
			if previousRow != 0 {
				break
			}

		}

		//for row := 0; row < len(matrix)-1; row++ {
		// Replace values
		//	0	 1    2	   3	4	 5	  6	   7	8
		// " ", "─", "│", "┌", "└", "┐", "┘", "▏", "▕"
		//if (col > 0 && row > 0) && matrix[row][col] == 1 {
		//	if matrix[row-1][col-1] == 1 {
		//		matrix[row][col] = 3 // "┌"
		//	} else if matrix[row+1][col-1] == 1 {
		//		matrix[row][col] = 4 // "└"
		//	} else if matrix[row-1][col+1] == 1 {
		//		matrix[row][col+1] = 5 // "┐"
		//	} else if matrix[row+1][col+1] == 1 {
		//		matrix[row][col+1] = 6 // "┘"
		//	}
		//}
		//	if (col > 0 && row > 0) && matrix[row][col] == 1 {
		//if (matrix[row][col-1] == 0 && matrix[row-1][col-1] == 1) &&
		//	(matrix[row][col+1] == 0 && matrix[row-1][col+1] == 1) {
		//	matrix[row][col-1] = 3
		//	matrix[row][col+1] = 5
		//}
		//		if matrix[row+1][col+1] == 1 {
		//			matrix[row][col] = 6
		//			matrix[row+1][col] = 3
		//		}
		//		if matrix[row-1][col-1] == 1 {
		//			matrix[row-1][col] = 4
		//			matrix[row][col+1] = 5
		//		}
		//	}
		//}
	}

	// build the matrix from the bottom->UP then left->right
	for row := len(matrix) - 1; row > 0; row-- {
		for col := 0; col < boxWidth-1; col++ {
			// Finally, start building the graph string using the matrix
			switch matrix[row][col] {
			case LineHorizontal:
				graph += string(LineHorizontal)
			case LineVertical:
				graph += string(LineVertical)
			case LineVerticalHeavy:
				graph += string(LineVerticalHeavy)
			case LineDownRight:
				graph += string(LineDownRight)
			case LineDownLeft:
				graph += string(LineDownLeft)
			case LineUpRight:
				graph += string(LineUpRight)
			case LineUpLeft:
				graph += string(LineUpLeft)
			case LineVerticalRight:
				graph += string(LineVerticalRight)
			case LineVerticalLeft:
				graph += string(LineVerticalLeft)
			case LineHorizontalUp:
				graph += string(LineHorizontalUp)
			case LineHorizontalDown:
				graph += string(LineHorizontalDown)
			case LineHorizontalVertical:
				graph += string(LineHorizontalVertical)
			default: // default accounts for case 0 and any other value not in the switch
				graph += " "
			}

			//graph += strconv.Itoa(matrix[row][col])

		}
		graph += "\n"
	}

	duration := time.Since(timestamp)
	if duration != 0 {
		timestampHistory = append(timestampHistory, duration)
		slog.Log(context.Background(), LevelPerf,
			"buildGraphMatrix() time: "+(duration).String()+", "+
				"mean time: "+strconv.FormatFloat(mean(timestampHistory), 'f', 2, 64)+"µs")
	}

	return graph
}

////##################################################################################////
////########################//// UI GOROUTINES START HERE ////########################////

// UpdateCPU is a text UI function that starts as a goroutine before the application
// starts.
func UpdateCPU(app *tview.Application, box *tview.TextView, showBorder bool) {
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

		stats := GetCPUStats()
		//lastIndex := len(stats) - 1

		//boxText = "CPU load: " + strconv.FormatFloat(
		//	stats[lastIndex].UsagePercent, 'f', 1, 64) + " %" + "\n"
		//boxText = "len of stats = " + strconv.Itoa(len(stats)) + "\n"
		//boxText = buildGraph(stats, width, height)

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}

		boxText = buildGraphMatrix(stats, width, height)
		app.QueueUpdateDraw(func() {
			box.SetText(boxText)
			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta = time.Since(timestamp) - *&Cfg.UpdateInterval
				}
				slog.Log(context.Background(), LevelPerf,
					"UpdateCPU() time: "+(timeDelta).String())
			}
		})

	}
}

func UpdateCPUTemp(app *tview.Application, box *tview.TextView, showBorder bool) {
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

		boxText, _ = CPUTemp()

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}
		app.QueueUpdateDraw(func() {
			box.SetText(boxText)
			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta = time.Since(timestamp) - *&Cfg.UpdateInterval
				}
				slog.Log(context.Background(), LevelPerf,
					"UpdateCPUTemp() time: "+timeDelta.String())
			}
		})

	}
}

//// Disk/HDD/SSD ////####################################################################

func UpdateDisk(app *tview.Application, box *tview.TextView, showBorder bool) {
	var (
		width, height int
		isResized     bool
		oldDisksStats []DiskStat
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(LblDisk)
	slog.Info("Starting `UpdateDisk()` UI goroutine ...")

	for {
		var boxText string

		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		disksStats = DisksStats()
		if oldDisksStats == nil {
			oldDisksStats = disksStats
		}

		for i, dsk := range disksStats {
			var vDiskStr string
			var diskCapacityStr string

			if dsk.IsVirtualDisk {
				vDiskStr = "<Virtual Disk>"
			}

			diskCapacity := ConvertBytesToGiB(dsk.Total, false)
			if diskCapacity < 999 {
				diskCapacityStr = strconv.FormatFloat(
					diskCapacity, 'f', 1, 64) + " GB"
			} else {
				diskCapacityStr = strconv.FormatFloat(
					diskCapacity/1000.0, 'f', 2, 64) + " TB"
			}

			boxText += buildBoxTitleStatRow(
				dsk.Mountpoint+" "+vDiskStr, diskCapacityStr, width, " ")

			// TODO: reflect disk size changes
			boxText += buildProgressBar(
				dsk.UsedPercent, oldDisksStats[i].UsedPercent, width, BLUE, WHITE)

			//boxText += "width=" + strconv.Itoa(width) + ", height=" + strconv.Itoa(height) + "\n"
		}

		// TODO: SCROLLING TEXT - get number of new lines in the box text and then compare
		// 	num of lines to height. Then, get the delta/difference between the height and
		// 	number of lines. Scroll/remove 1 line at a time per second, and reset back to
		// 	0 when the last line is reached

		oldDisksStats = disksStats

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}
		app.QueueUpdateDraw(func() {
			box.SetText(boxText)
			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta = time.Since(timestamp) - *&Cfg.UpdateInterval
				}
				slog.Log(context.Background(), LevelPerf,
					"UpdateDisk() time: "+timeDelta.String())
			}
		})

	}
}

//// GPU ////#############################################################################

func UpdateGPU(app *tview.Application, box *tview.TextView, showBorder bool) {
	var (
		boxText       string
		width, height int
		isResized     bool

		oldGPUStats   []GPUStat
		oldGPULoadBar string
		oldGPUMemBar  string
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(" " + GPUName() + " ")
	slog.Info("Starting `UpdateGPU()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		gpuStats = GPUStats()
		if oldGPUStats == nil {
			oldGPUStats = gpuStats
		}

		lastElement := len(gpuStats) - 1
		oldLastElement := len(oldGPUStats) - 1
		/// END DATA FETCH

		gpuLoadStr := strconv.FormatInt(int64(gpuStats[lastElement].Load*100.0), 10) + "%"
		gpuLoadTitleRow := buildBoxTitleStatRow("Load:", gpuLoadStr, width, " ")

		gpuMemoryUsageRatio := gpuStats[lastElement].MemoryUsage / gpuStats[lastElement].MemoryTotal
		oldGPUMemoryUsageRatio :=
			oldGPUStats[oldLastElement].MemoryUsage / oldGPUStats[oldLastElement].MemoryTotal
		gpuMemoryStr := strconv.FormatInt(int64(gpuMemoryUsageRatio*100), 10) + "%"
		gpuMemoryTitleRow := buildBoxTitleStatRow("Mem:", gpuMemoryStr, width, " ")

		boxText = gpuLoadTitleRow
		gpuLoadBar := buildProgressBar(
			gpuStats[lastElement].Load, oldGPUStats[oldLastElement].Load, width, GREEN, WHITE)
		if oldGPULoadBar == "" || oldGPULoadBar == gpuLoadBar {
			oldGPULoadBar = gpuLoadBar
			boxText += gpuLoadBar
		} else {
			boxText += gpuLoadBar
		}

		boxText += "\n" // add an extra line gap to visually and obviously separate the info

		boxText += gpuMemoryTitleRow
		gpuMemBar := buildProgressBar(
			gpuMemoryUsageRatio, oldGPUMemoryUsageRatio, width, GREEN, WHITE)
		if oldGPUMemBar == "" || oldGPUMemBar == gpuMemBar {
			oldGPUMemBar = gpuMemBar
			boxText += gpuMemBar
		} else {
			boxText += gpuMemBar
		}

		oldGPUStats = gpuStats

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}
		app.QueueUpdateDraw(func() {
			box.SetText(boxText)
			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta = time.Since(timestamp) - *&Cfg.UpdateInterval
				}
				slog.Log(context.Background(), LevelPerf,
					"UpdateGPU() time: "+timeDelta.String())
			}
		})
	}
}

func UpdateGPUTemp(app *tview.Application, box *tview.TextView, showBorder bool) {
	var (
		boxText       string
		width, height int
		isResized     bool
		oldGPUStats   []GPUStat
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(LblGPUTemp)
	slog.Info("Starting `UpdateGPUTemp()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		gpuStats = GPUStats()
		if len(oldGPUStats) == 0 {
			oldGPUStats = gpuStats
		}
		lastElement := len(gpuStats) - 1
		oldLastElement := len(oldGPUStats) - 1

		//boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		gpuTempStr := strconv.Itoa(int(gpuStats[lastElement].Temperature)) + "°C"
		gpuTempTitle := buildBoxTitleStatRow("Temp:", gpuTempStr, width, " ")

		gpuTempRatio := float64(gpuStats[lastElement].Temperature) / 100.0
		oldGPUTempRatio := float64(oldGPUStats[oldLastElement].Temperature / 100.0)

		boxText = gpuTempTitle + buildProgressBar(gpuTempRatio, oldGPUTempRatio, width, GREEN, WHITE)

		oldGPUStats = gpuStats

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}
		app.QueueUpdateDraw(func() {
			box.SetText(boxText)
			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta = time.Since(timestamp) - *&Cfg.UpdateInterval
				}
				slog.Log(context.Background(), LevelPerf,
					"UpdateGPUTemp() time: "+timeDelta.String())
			}
		})
	}
}

//// Memory ////##########################################################################

func UpdateMemory(app *tview.Application, box *tview.TextView, showBorder bool) {
	var (
		boxText       string
		width, height int
		isResized     bool
		memData       []*mem.VirtualMemoryStat
		oldMemStats   *mem.VirtualMemoryStat
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(LblMemory)
	slog.Info("Starting `UpdateMemory()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		memStats = MemoryStats()
		memData = append(memData, memStats)
		if oldMemStats == nil {
			oldMemStats = memStats
		}
		/// END DATA FETCH

		memUsed := ConvertBytesToGiB(memStats.Used, false)
		memUsedText := strconv.FormatFloat(memUsed, 'f', 1, 64) + " GB"

		memTotal := ConvertBytesToGiB(memStats.Total, false)
		memTotalText := strconv.FormatFloat(memTotal, 'f', 1, 64) + " GB"

		memoryUsedTitleRow := buildBoxTitleStatRow("Used", "Total", width, " ")
		progressBar := buildProgressBar(
			memStats.UsedPercent/100, oldMemStats.UsedPercent/100, width, GREEN, WHITE)
		memoryStatsRow := buildBoxTitleStatRow(memUsedText, memTotalText, width, " ")

		boxText = memoryUsedTitleRow + progressBar + memoryStatsRow

		oldMemStats = memStats

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}
		app.QueueUpdateDraw(func() {
			box.SetText(boxText)
			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta = time.Since(timestamp) - *&Cfg.UpdateInterval
				}
				slog.Log(context.Background(), LevelPerf,
					"UpdateMemory() time: "+timeDelta.String())
			}
		})
	}
}

//// Network ////#########################################################################

func UpdateNetwork(app *tview.Application, box *tview.TextView, showBorder bool) {
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

		netInfo = NetworkStats()

		boxText = buildBoxTitleCentered("//"+Hostname(), RED, width, " ")
		//boxText += "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		for _, iface := range netInfo {
			boxText += buildBoxTitleStatRow(
				"DOWN: ", strconv.FormatUint(iface.BytesSent, 10), width, " ")
			boxText += buildBoxTitleStatRow(
				"UP: ", strconv.FormatUint(iface.BytesRecv, 10), width, " ")
		}

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}
		app.QueueUpdateDraw(func() {
			box.SetText(boxText)
			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta = time.Since(timestamp) - *&Cfg.UpdateInterval
				}
				slog.Log(context.Background(), LevelPerf,
					"UpdateNetwork() time: "+timeDelta.String())
			}
		})
	}
}

//// Processes ////#######################################################################

func UpdateProcesses(app *tview.Application, box *tview.Table, showBorder bool) {
	var (
		//boxText       string
		width, height int
		isResized     bool
	)

	box.SetBorder(showBorder).SetTitle(LblProc)
	slog.Info("Starting `UpdateProcesses()` UI goroutine ...")

	for {
		timestamp := time.Now()
		// TODO: Get process info here then pass it into the app.QueueUpdateDraw()
		// 	before sleeping

		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}
		app.QueueUpdateDraw(func() {
			//box.SetText(boxText)
			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta = time.Since(timestamp) - *&Cfg.UpdateInterval
				}
				slog.Log(context.Background(), LevelPerf,
					"UpdateProcesses() time: "+timeDelta.String())
			}
		})
	}
}
