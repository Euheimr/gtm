package ui

import (
	"github.com/rivo/tview"
	"gtm"
	"log/slog"
	"math"
	"runtime"
	"strconv"
	"strings"
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

// these constants are text formatting tags used by the tcell package
const (
	BLACK  string = "[black]"
	BLUE          = "[blue]"
	GREEN         = "[green]"
	GRAY          = "[gray]"
	RED           = "[red]"
	WHITE         = "[white]"
	YELLOW        = "[yellow]"
)

var barSymbols = [8]string{" ", "░", "▒", "▓", "█", "[", "|", "]"}

const (
	LblDisk    = " HDD / SSD "
	LblCPUTemp = " CPU Temp "
	LblGPUTemp = " GPU Temp "
	LblMemory  = " Memory "
	LblNetwork = " Network "
	LblProc    = " Processes "
)

var (
	Layout layoutMain
	LblCPU = " CPU "
	LblGPU = " GPU "
)

var Cfg = &gtm.Cfg

func init() {
	// Initialize the main Layout ASAP
	Layout = layoutMain{
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
	cpuParentBox.SetBorder(true).SetTitle(" " + gtm.GetCPUModel(true) + " ")
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
	if Cfg.EnableGPU && runtime.GOOS != "darwin" {
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

func SleepWithTimestampDelta(timestamp time.Time, update time.Duration) {
	timeDelta := time.Now().UnixMilli() - timestamp.UnixMilli()
	if timeDelta < update.Milliseconds() {
		time.Sleep(time.Duration(update.Milliseconds() - timeDelta))
	}
}

func BuildProgressBar(ratio float64, columns int, colorFill string, colorEmpty string) string {
	var (
		barUsed       int    = 0
		spacingOffset        = columns   // default the spacing offset to box columns
		barText       string = colorFill // insert "used" color tag here
		charUsed             = barSymbols[4]
		charEmpty            = barSymbols[1]
		charStart            = barSymbols[4]
		charEnd              = barSymbols[1]
	)

	if ratio <= 1.0 {
		// We never want a ratio higher than 1.0 or the bar will overflow to the next line
		barUsed = int(math.Round(float64(columns) * ratio))
	} else {
		barUsed = int(math.Round(float64(columns) * 1.0)) // Clamp the ratio to 1.0
	}

	if barUsed >= 1 {
		barText += charStart
		for i := range barUsed {
			// If we aren't on the last element, build a bar of "used memory"
			if i != (barUsed - 1) {
				barText += charUsed
			}
		}
		spacingOffset -= barUsed
	}
	// Add in the second color tag for the empty or "unused" portion of the bar
	barText += colorEmpty

	for i := 0; i < (spacingOffset - 1); i++ {
		// Iterate over the spacing offset to fill in the empty/unused part of the bar
		barText += charEmpty
	}
	barText += charEnd // Cap off the end of the bar
	return barText + WHITE + "\n"
}

func GetInnerBoxSize(box *tview.Box, oldWidth int, oldHeight int) (width int, height int,
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

func InsertCenterSpacing(arg1 string, arg2 string, boxWidth int,
	spaceChar string) (spaces string) {

	spacingCount := boxWidth - len(arg1) - len(arg2)
	for range spacingCount {
		spaces += spaceChar
	}
	return spaces
}

func BuildBoxTitleRow(title string, statStr string, boxWidth int, spaceChar string) string {
	return title + InsertCenterSpacing(title, statStr, boxWidth, spaceChar) + statStr + "\n"
}
