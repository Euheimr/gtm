package ui

import (
	"github.com/rivo/tview"
	"gtm"
	"log/slog"
	"strconv"
	"time"
)

var diskInfo []gtm.DiskInfo

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
		width, height, _ = GetInnerBoxSize(Layout.Disk.Box, width, height)

		diskInfo = gtm.GetDiskInfo()
		boxText = ""

		for _, dsk := range diskInfo {
			var diskCapacityStr string
			diskCapacity := gtm.ConvertBytesToGiB(dsk.Total, false)
			if diskCapacity < 999 {
				diskCapacityStr = strconv.FormatFloat(diskCapacity, 'f', 1, 64) + "GB"
			} else {
				diskCapacityStr = strconv.FormatFloat(diskCapacity/100.0, 'f', 1, 64) + "TB"
			}

			//boxText += dsk.Mountpoint + " | " + strconv.FormatBool(dsk.IsVirtualDisk) +
			//	" | " + strconv.FormatFloat(dsk.UsedPercent, 'g', -1, 64) +
			//	"% of " + diskCapacityStr + "\n"

			boxText += BuildBoxTitleRow(dsk.Mountpoint, diskCapacityStr, width, " ")
			boxText += BuildProgressBar(dsk.UsedPercent, width, RED, WHITE)
		}

		SleepWithTimestampDelta(timestamp, update, isResized)

		app.QueueUpdateDraw(func() {
			Layout.Disk.SetText(boxText)
		})
	}
}
