package gtm

import (
	"context"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v4/mem"
)

// These constants are text formatting tags used by the tcell package.
const (
	BLACK  string = "[black]"
	BLUE   string = "[blue]"
	GREEN  string = "[green]"
	GRAY   string = "[gray]"
	RED    string = "[red]"
	WHITE  string = "[white]"
	YELLOW string = "[yellow]"
)

const (
	LblCPUTemp string = " CPU Temp "
	LblDisk    string = " HDD / SSD "
	LblGPUTemp string = " GPU Temp "
	LblMemory  string = " Memory "
	LblNetwork string = " Network "
	LblProc    string = " Processes "
)

var (
	barSymbols = [5]string{" ", "░", "▒", "▓", "█"}
	// ascii codes => https://theasciicode.com.ar/
	// lineSymbols = [9]string{" ", "─", "│", "┌", "└", "┐", "┘", "▏", "▕"}
	// treeSymbols      = [5]string{"│", "┤", "├", "─", "└"}.
	blockSymbols = [4]string{"█", "▄", "■", "▀"}
	// directionSymbols = [8]string{"↑", "↓", "←", "→", "↖", "↗", "↘", "↙"}.
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

// Calculates the delta between the timestamp parameter and the current time and sleeps using that
// result. If the delta time is greater than the configured UpdateInterval, this function returns
// immediately instead of sleeping.
func sleepWithTimestampDelta(timestamp time.Time) {
	timeDelta := time.Now().UnixMilli() - timestamp.UnixMilli()

	switch {
	case timeDelta < Cfg.UpdateInterval.Milliseconds():
		// slog.Debug("sleep timeDelta = " + strconv.Itoa(int(update.Milliseconds()-timeDelta)))
		time.Sleep(time.Duration(Cfg.UpdateInterval.Milliseconds() - timeDelta))
	case timeDelta > Cfg.UpdateInterval.Milliseconds():
		// the timeDelta is greater than the update, don't sleep and update immediately
		time.Sleep(0)
	default:
		// slog.Debug("sleep update = " + strconv.Itoa(int(update.Milliseconds())))
		time.Sleep(Cfg.UpdateInterval)
	}
}

func convertMatrixToString(matrix [][]string) string {
	var text strings.Builder

	for row := range matrix {
		for col := 0; col <= len(matrix[row]); col++ {
			if col != len(matrix[row]) {
				text.WriteString(matrix[row][col])
			}
		}

		text.WriteString("\n")
	}

	return text.String()
}

func buildUtilBarVertical(ratio float64, rows int) string {
	// oldRatio float64,
	// colorFill string,
	var (
		countFill int
		// oldCountFill int
		charUsed = blockSymbols[0]
	)

	bar := make([][]string, rows)
	for i := range bar {
		bar[i] = make([]string, 2)
	}

	// If the ratio is greater than 1.0, then a programming error has occurred and I need
	//	to know about it ASAP
	if ratio > 1.0 || ratio < 0.0 {
		panic("progress bar ratio cannot be greater than 1.0 or less than 0.0 !")
	}

	countFill = int(math.Round(float64(rows) * ratio))
	// oldCountFill = int(math.Round(float64(rows) * oldRatio))

	if countFill >= 1 {
		countEmpty := rows - countFill

		for char := range rows {
			if char <= countEmpty {
				bar[char][0] = " "
				bar[char][1] = " "
			} else {
				bar[char][0] = charUsed
				bar[char][1] = charUsed
			}
		}
	}

	return convertMatrixToString(bar)
}

func buildUtilBarHorizontal(ratio float64, oldRatio float64, columns int, colorFill string) string {
	var (
		countFill    int
		oldCountFill int
		countEmpty   = columns // default char count to total box columns (box width)
		charUsed     = barSymbols[4]
		charOld      = barSymbols[2]
		charEmpty    = barSymbols[1]
		charStart    = barSymbols[4]
		colorEmpty   = WHITE
	)
	// If the ratio is greater than 1.0, then a programming error has occurred and I need
	//	to know about it ASAP
	if ratio > 1.0 || ratio < 0.0 {
		panic("progress bar ratio cannot be greater than 1.0 or less than 0.0 !")
	}

	var barText strings.Builder

	// Color the bar based on how full they are
	switch {
	case ratio >= 0.85:
		barText.WriteString(RED)
	case ratio >= 0.7 && ratio < 0.85:
		barText.WriteString(YELLOW)
	default:
		barText.WriteString(colorFill)
	}

	countFill = int(math.Round(float64(columns) * ratio))
	oldCountFill = int(math.Round(float64(columns) * oldRatio))

	if countFill >= 1 {
		barText.WriteString(charStart)

		for i := range countFill {
			if i != (countFill - 1) {
				// If we aren't on the last element, build a bar of "used" memory
				barText.WriteString(charUsed)
			}
		}

		countEmpty -= countFill
	} else if ratio > 0.0 {
		// If we are above 0% load/usage, then always show at least 1 character indicating
		//	"used" or a "load"
		barText.WriteString(charUsed)
		// Also, we need to make sure the countEmpty is -1 to not overflow text to the next
		//	line
		countEmpty -= 1
	}

	if oldCountFill > countFill {
		barText.WriteString(RED)

		positiveDelta := oldCountFill - countFill
		for range positiveDelta {
			barText.WriteString(charOld)
		}

		countEmpty -= positiveDelta
	}
	// Add in the second color tag for the "empty" or "unused" portion of the bar
	barText.WriteString(colorEmpty)

	// Iterate over an integer count of empty chars to add in the empty/unused part of
	//	the bar.
	for range countEmpty {
		barText.WriteString(charEmpty)
	}

	return barText.String()
}

// getInnerBoxSizeReturns the box width, height and isResized is true if the current box size does
// not match oldWidth or oldHeight parameters.
//
// Uses the box primitive to get the width and height (number of columns and rows respectively) and
// compares that result with the oldWidth and oldHeight parameters.
func getInnerBoxSize(box *tview.Box, oldWidth int, oldHeight int) (int, int, bool) {
	_, _, width, height := box.GetInnerRect()
	isResized := false

	if (oldWidth != 0 || oldHeight != 0) && (oldWidth != width || oldHeight != height) {
		isResized = true

		if Cfg.PerformanceLoggingUI {
			boxTitle := strings.TrimSpace(box.GetTitle())
			slog.Debug(boxTitle + " inner box size changed from (" +
				strconv.Itoa(oldWidth) + "->" + strconv.Itoa(width) + ") columns " +
				"and (" + strconv.Itoa(oldHeight) + "->" + strconv.Itoa(height) + ") rows !")
		}
	}

	return width, height, isResized
}

func buildBoxTitleStatRow(title string, statStr string, boxWidth int) string {
	var (
		spaces       strings.Builder
		spacingCount = boxWidth - len(title) - len(statStr)
	)

	for range spacingCount {
		spaces.WriteString(" ")
	}

	return title + spaces.String() + statStr + "\n"
}

func buildBoxTitleCentered(title string, color string, boxWidth int) string {
	var titleString strings.Builder

	spacesCount := boxWidth - len(title)
	for range spacesCount / 2 {
		titleString.WriteString(" ")
	}

	titleString.WriteString(color + title + WHITE)

	for range spacesCount / 2 {
		titleString.WriteString(" ")
	}

	return titleString.String() + "\n"
}

func mean(timestamps []time.Duration) float64 {
	total := 0.0

	for _, t := range timestamps {
		total += float64(t.Microseconds())
	}

	return total / float64(len(timestamps))
}

func buildGraphMatrix(stat any, boxWidth int, boxHeight int) string {
	var graph strings.Builder

	timestamp := time.Now()

	// The first index of the matrix is the height, the second index is the width and the
	//	value at [x][y] equals the index of the symbol to use for the graph string
	matrix := make([][]int32, boxHeight)
	for row := range matrix {
		matrix[row] = make([]int32, boxWidth)
	}

	switch stat := stat.(type) {
	case []CPUStat:
		// make sure we just take the last n stats for the matrix so we don't overflow index
		if len(stat) > boxWidth {
			stat = stat[len(stat)-boxWidth:]
		}

		var rows []int

		for key, val := range stat {
			row := int(math.Round((val.UsagePercent / 100) * float64(boxHeight)))
			if row != 0 {
				rows = append(rows, row)
				matrix[row][key] = LineHorizontal
			}
		}

		// lastIndex := len(rows) - 1

		// for col, row := range rows {
		//	if col > 0 && col != lastIndex && col+1 != lastIndex {
		//		if row > rows[col+1] {
		//			// insert vertical lines
		//			//rowDelta := row - rows[col+1] - 1
		//			//for r := range rowDelta {
		//			//	matrix[r+rowDelta+1][col] = LineVertical
		//			//}
		//			//if matrix[row-1][col] == LineVertical {
		//			//	matrix[row][col] = LineDownLeft
		//			//}
		//			//if matrix[row][col] == LineDownLeft {
		//			//	matrix[rowDelta][col] = LineUpRight
		//			//}
		//		} else if row < rows[col+1] {
		//			//rowDelta := rows[col+1] - row - 1
		//			//for r := range rowDelta {
		//			//	matrix[r+rowDelta+1][col+1] = LineVertical
		//			//	if r == rowDelta-1 {
		//			//		matrix[rows[col+1]][col+1] = LineDownRight
		//			//	}
		//			//}
		//			//if matrix[row+1][col+1] == LineVertical {
		//			//	matrix[row][col+1] = LineUpLeft
		//			//}
		//			//matrix[rows[col+1]][col+1] = LineDownRight
		//			//if matrix[row+1][col] == LineVertical {
		//			//	matrix[row][col] = LineUpLeft
		//			//}
		//		}
		//	}
		//}

		var rowsStr strings.Builder
		for _, row := range rows {
			rowsStr.WriteString(strconv.Itoa(row))
		}
		// slog.Debug("rowsStr = " + rowsStr)

	case []CPUTempStat:
	case []DiskStat:
	case []GPUStat:
	}

	// build the matrix from the bottom->UP then left->right
	for row := len(matrix) - 1; row >= 0; row-- {
		for col := range boxWidth {
			// Finally, start building the graph string using the matrix
			switch matrix[row][col] {
			case LineHorizontal:
				graph.WriteString(string(LineHorizontal))
			case LineVertical:
				graph.WriteString(string(LineVertical))
			case LineVerticalHeavy:
				graph.WriteString(string(LineVerticalHeavy))
			case LineDownRight:
				graph.WriteString(string(LineDownRight))
			case LineDownLeft:
				graph.WriteString(string(LineDownLeft))
			case LineUpRight:
				graph.WriteString(string(LineUpRight))
			case LineUpLeft:
				graph.WriteString(string(LineUpLeft))
			case LineVerticalRight:
				graph.WriteString(string(LineVerticalRight))
			case LineVerticalLeft:
				graph.WriteString(string(LineVerticalLeft))
			case LineHorizontalUp:
				graph.WriteString(string(LineHorizontalUp))
			case LineHorizontalDown:
				graph.WriteString(string(LineHorizontalDown))
			case LineHorizontalVertical:
				graph.WriteString(string(LineHorizontalVertical))
			default: // default accounts for case 0 and any other value not in the switch
				graph.WriteString(" ") // "·"
			}

			// graph.WriteString(strconv.Itoa(matrix[row][col]))
		}

		graph.WriteString("\n")
	}

	// if Cfg.Debug {
	//	slog.Debug("buildGraphMatrix():\n" + graph)
	//}

	duration := time.Since(timestamp)
	timestampHistory = append(timestampHistory, duration)
	slog.Log(context.Background(), LevelPerf,
		"buildGraphMatrix() time: "+(duration).String()+", "+
			"mean time: "+strconv.FormatFloat(mean(timestampHistory), 'f', 2, 64)+"µs")

	return graph.String()
}

////##################################################################################////
////########################//// UI GOROUTINES START HERE ////########################////

// UpdateCPU is a text UI function that starts as a goroutine before the application
// starts.
func UpdateCPU(app *tview.Application, box *tview.TextView, showBorder bool) {
	var (
		boxText       strings.Builder
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
		// boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height) + "\n"

		stats := GetCPUStats()
		// boxText = buildGraph(stats, width, height)

		// TEST DATA!
		// stats := []CPUStat{
		//	0:  {20.0},
		//	1:  {25.0},
		//	2:  {30.0},
		//	3:  {40.0},
		//	4:  {40.0},
		//	5:  {40.0},
		//	6:  {20.0},
		//	7:  {40.0},
		//	8:  {40.0},
		//	9:  {60.0},
		//	10: {10.0},
		//	11: {45.0},
		//	12: {20.0},
		//	13: {20.0},
		//	14: {60.0},
		//	15: {50.0},
		//	16: {50.0},
		//	17: {50.0},
		//}

		boxText.Reset()
		boxText.WriteString(buildGraphMatrix(stats, width, height))

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}

		app.QueueUpdateDraw(func() {
			box.SetText(boxText.String())

			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta -= Cfg.UpdateInterval
				}

				slog.Log(context.Background(), LevelPerf,
					"UpdateCPU() time: "+(timeDelta).String())
			}
		})
	}
}

func UpdateCPUTemp(app *tview.Application, box *tview.TextView, showBorder bool) {
	var (
		boxText       strings.Builder
		width, height int
		isResized     bool
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(LblCPUTemp)
	slog.Info("Starting `UpdateCPUTemp()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		// boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height) + "\n"

		temp, _ := CPUTemp()

		boxText.Reset()
		boxText.WriteString(temp)

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}

		app.QueueUpdateDraw(func() {
			box.SetText(boxText.String())

			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta -= Cfg.UpdateInterval
				}

				slog.Log(context.Background(), LevelPerf,
					"UpdateCPUTemp() time: "+timeDelta.String())
			}
		})
	}
}

func UpdateDisk(app *tview.Application, box *tview.TextView, showBorder bool) {
	var (
		boxText       strings.Builder
		width, height int
		isResized     bool
		oldDisksStats []DiskStat
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(LblDisk)
	slog.Info("Starting `UpdateDisk()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		disksStats = DisksStats()
		if oldDisksStats == nil {
			oldDisksStats = disksStats
		}

		boxText.Reset()

		for i, dsk := range disksStats {
			var (
				vDiskStr        string
				diskCapacityStr string
			)

			if dsk.IsVirtualDisk {
				vDiskStr = "<VDisk>"
			}

			diskCapacity := ConvertBytesToGiB(dsk.Total, false)
			if diskCapacity < 999 {
				diskCapacityStr = strconv.FormatFloat(
					diskCapacity, 'f', 1, 64) + " GB"
			} else {
				diskCapacityStr = strconv.FormatFloat(
					diskCapacity/1000.0, 'f', 2, 64) + " TB"
			}

			boxText.WriteString(buildBoxTitleStatRow(
				dsk.Mountpoint+" "+vDiskStr, diskCapacityStr, width))

			// TODO: reflect disk size changes
			boxText.WriteString(buildUtilBarHorizontal(
				dsk.UsedPercent, oldDisksStats[i].UsedPercent, width, BLUE))

			// boxText += "width=" + strconv.Itoa(width) + ", height=" + strconv.Itoa(height) + "\n"
		}

		// TODO: SCROLLING TEXT - get number of new lines in the box text and then compare
		// 	num of lines to height. Then, get the delta/difference between the height and
		// 	number of lines. Scroll/remove 1 line at a time per second, and reset back to
		// 	0 when the last line is reached
		// TODO: OR we use vertical bars??
		// boxText += buildUtilizationBarVertical(disksStats[0].UsedPercent,
		//  oldDisksStats[0].UsedPercent, height, BLUE, WHITE)

		oldDisksStats = disksStats

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}

		app.QueueUpdateDraw(func() {
			box.SetText(boxText.String())

			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta -= Cfg.UpdateInterval
				}

				slog.Log(context.Background(), LevelPerf,
					"UpdateDisk() time: "+timeDelta.String())
			}
		})
	}
}

func UpdateGPU(app *tview.Application, box *tview.TextView, showBorder bool) {
	var (
		boxText       strings.Builder
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
		gpuLoadTitleRow := buildBoxTitleStatRow("Load:", gpuLoadStr, width)

		gpuMemoryUsageRatio := gpuStats[lastElement].MemoryUsage /
			gpuStats[lastElement].MemoryTotal
		oldGPUMemoryUsageRatio := oldGPUStats[oldLastElement].MemoryUsage /
			oldGPUStats[oldLastElement].MemoryTotal

		gpuMemoryStr := strconv.FormatInt(int64(gpuMemoryUsageRatio*100), 10) + "%"
		gpuMemoryTitleRow := buildBoxTitleStatRow("Mem:", gpuMemoryStr, width)

		boxText.Reset()
		boxText.WriteString(gpuLoadTitleRow)

		gpuLoadBar := buildUtilBarHorizontal(
			gpuStats[lastElement].Load, oldGPUStats[oldLastElement].Load, width, GREEN)
		if oldGPULoadBar == "" || oldGPULoadBar == gpuLoadBar {
			oldGPULoadBar = gpuLoadBar
			boxText.WriteString(gpuLoadBar)
		} else {
			boxText.WriteString(gpuLoadBar)
		}

		// add an extra line gap to visually and obviously separate the info
		boxText.WriteString("\n")

		boxText.WriteString(gpuMemoryTitleRow)

		gpuMemBar := buildUtilBarHorizontal(
			gpuMemoryUsageRatio, oldGPUMemoryUsageRatio, width, GREEN)
		if oldGPUMemBar == "" || oldGPUMemBar == gpuMemBar {
			oldGPUMemBar = gpuMemBar
			boxText.WriteString(gpuMemBar)
		} else {
			boxText.WriteString(gpuMemBar)
		}

		// slog.Debug("UPDATE_GPU: \n" + boxText)

		oldGPUStats = gpuStats

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}

		app.QueueUpdateDraw(func() {
			box.SetText(boxText.String())

			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta -= Cfg.UpdateInterval
				}

				slog.Log(context.Background(), LevelPerf,
					"UpdateGPU() time: "+timeDelta.String())
			}
		})
	}
}

func UpdateGPUTemp(app *tview.Application, box *tview.TextView, showBorder bool) {
	var (
		boxText       strings.Builder
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

		// boxText = "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		gpuTempStr := strconv.Itoa(int(gpuStats[lastElement].Temperature)) + "°C"
		gpuTempTitle := buildBoxTitleStatRow("Temp:", gpuTempStr, width)

		gpuTempRatio := float64(gpuStats[lastElement].Temperature) / 100.0
		oldGPUTempRatio := float64(oldGPUStats[oldLastElement].Temperature / 100.0)

		boxText.Reset()
		boxText.WriteString(gpuTempTitle + buildUtilBarHorizontal(
			gpuTempRatio,
			oldGPUTempRatio,
			width,
			GREEN,
		))

		oldGPUStats = gpuStats

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}

		app.QueueUpdateDraw(func() {
			box.SetText(boxText.String())

			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta -= Cfg.UpdateInterval
				}

				slog.Log(context.Background(), LevelPerf,
					"UpdateGPUTemp() time: "+timeDelta.String())
			}
		})
	}
}

func UpdateMemory(app *tview.Application, box *tview.TextView, showBorder bool) {
	var (
		boxText       strings.Builder
		width, height int
		isResized     bool
		// memData       []*mem.VirtualMemoryStat
		oldMemStats *mem.VirtualMemoryStat
	)

	box.SetDynamicColors(true)
	box.SetBorder(showBorder).SetTitle(LblMemory)
	slog.Info("Starting `UpdateMemory()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		memStats = MemoryStats()

		// memData = append(memData, memStats)
		if oldMemStats == nil {
			oldMemStats = memStats
		}
		/// END DATA FETCH

		memUsed := ConvertBytesToGiB(memStats.Used, false)
		memUsedText := strconv.FormatFloat(memUsed, 'f', 1, 64) + " GB"

		memTotal := ConvertBytesToGiB(memStats.Total, false)
		memTotalText := strconv.FormatFloat(memTotal, 'f', 1, 64) + " GB"

		memoryUsedTitleRow := buildBoxTitleStatRow("Used", "Total", width)
		progressBar := buildUtilBarHorizontal(
			memStats.UsedPercent/100, oldMemStats.UsedPercent/100, width, GREEN)
		memoryStatsRow := buildBoxTitleStatRow(memUsedText, memTotalText, width)

		boxText.Reset()
		boxText.WriteString(memoryUsedTitleRow + progressBar + memoryStatsRow)

		oldMemStats = memStats

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}

		app.QueueUpdateDraw(func() {
			box.SetText(boxText.String())

			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta -= Cfg.UpdateInterval
				}

				slog.Log(context.Background(), LevelPerf,
					"UpdateMemory() time: "+timeDelta.String())
			}
		})
	}
}

func UpdateNetwork(app *tview.Application, box *tview.TextView, showBorder bool) {
	var (
		boxText       strings.Builder
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

		boxText.Reset()
		boxText.WriteString(buildBoxTitleCentered("//"+Hostname(), RED, width))
		// boxText += "col: " + strconv.Itoa(width) + ", row: " + strconv.Itoa(height)
		for _, iface := range netInfo {
			boxText.WriteString(
				buildBoxTitleStatRow("DOWN: ", strconv.FormatUint(iface.BytesSent, 10), width))
			boxText.WriteString(
				buildBoxTitleStatRow("UP: ", strconv.FormatUint(iface.BytesRecv, 10), width))
		}

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}

		app.QueueUpdateDraw(func() {
			box.SetText(boxText.String())

			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta -= Cfg.UpdateInterval
				}

				slog.Log(context.Background(), LevelPerf,
					"UpdateNetwork() time: "+timeDelta.String())
			}
		})
	}
}

func UpdateProcesses(app *tview.Application, box *tview.Table, showBorder bool) {
	var (
		// boxText       string
		width, height int
		isResized     bool
	)

	box.SetBorder(showBorder).SetTitle(LblProc)
	slog.Info("Starting `UpdateProcesses()` UI goroutine ...")

	for {
		timestamp := time.Now()
		// TODO: Get process info here then pass it into the app.QueueUpdateDraw() before sleeping

		width, height, isResized = getInnerBoxSize(box.Box, width, height)

		if !isResized {
			sleepWithTimestampDelta(timestamp)
		}

		app.QueueUpdateDraw(func() {
			// box.SetText(boxText)
			if Cfg.PerformanceLoggingUI {
				timeDelta := time.Since(timestamp)
				if isResized {
					timeDelta -= Cfg.UpdateInterval
				}

				slog.Log(context.Background(), LevelPerf,
					"UpdateProcesses() time: "+timeDelta.String())
			}
		})
	}
}
