package ui

import (
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v4/disk"
	"log/slog"
	"strconv"
	"time"
)

var (
	diskInfo  []disk.PartitionStat
	diskUsage []disk.UsageStat
)

func UpdateDisk(app *tview.Application, showBorder bool, update time.Duration) {
	var (
		boxText         string
		width, height   int
		isResized       bool
		disksVirtualStr []bool
	)

	Layout.Disk.SetDynamicColors(true)
	Layout.Disk.SetBorder(showBorder).SetTitle(LblDisk)
	slog.Info("Starting `UpdateDisk()` UI goroutine ...")

	for {
		timestamp := time.Now()
		width, height, _ = GetInnerBoxSize(Layout.Disk.Box, width, height)

		for i, dsk := range disksVirtualStr {
			boxText += diskInfo[i].Device + " is RAMDISK=" + strconv.FormatBool(dsk) + "\n"
		}

		SleepWithTimestampDelta(timestamp, update, isResized)

		app.QueueUpdateDraw(func() {
			Layout.Disk.SetText(boxText)
		})
	}
}
