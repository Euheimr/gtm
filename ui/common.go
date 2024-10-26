package ui

import (
	"github.com/rivo/tview"
	"gtm"
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
	Cfg         = &gtm.Cfg
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
	if gtm.HasGPU() {
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

func SleepWithTimestampDelta(timestamp time.Time, update time.Duration, isResized bool) {
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

func BuildProgressBar(ratio float64, columns int, colorFill string, colorEmpty string) string {
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
