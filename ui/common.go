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
	RED           = "[red]"
	WHITE         = "[white]"
	YELLOW        = "[yellow]"
	GRAY          = "[gray]"
)

var blockSymbols = [4]string{"░", "▒", "▓", "█"}

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
	cpuParentBox.SetBorder(true).SetTitle(" " + gtm.GetCpuModel() + " ")
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

func BuildHorizontalTextBar(columns int, ratio float64, color1 string, color2 string) string {
	// FIXME: rename the color1 / color2 params?
	var (
		startChar = blockSymbols[3]
		usedChar  = blockSymbols[3]
		spaceChar = blockSymbols[0]
		endChar   = blockSymbols[0]
	)
	// if we are color coding, insert the color code tag here for USED
	barText := color1
	barText += startChar
	// We -1 the size of the "used" part of the bar to make room for the first element
	//	containing the "[" bracket
	barUsed := int(math.Round(float64(columns)*ratio)) - 1
	// The spacing offset is the inner width of the box minus the characters representing
	//	total memory (the bar "|" characters). We -1 here to make room for the matching
	//	closing "]" bracket (or the last character) for a complete bar
	spacingOffset := columns - barUsed - 1

	for i := range barUsed {

		// If we aren't on the last element, builds a bar of "used memory"
		//	ie.  [|||||||    <- like this
		if i != (barUsed - 1) {
			barText += usedChar
		} else {
			// Add in the second color for the "unused" portion of the bar
			barText += color2
			for range spacingOffset {
				// [||||||||        ] Now add the spacing offset to make a complete bar
				//         HERE ^
				barText += spaceChar
			}
			barText += endChar
			// the complete bar should look like:  [||||||||        ] (at 50% load)
		}
	}
	return barText
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
		spaces = spaces + spaceChar
	}
	return spaces
}
