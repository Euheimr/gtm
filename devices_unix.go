//go:build linux || darwin

package gtm

import (
	"log/slog"
	"os"

	"github.com/shirou/gopsutil/v4/disk"
)

func isAdmin(elevated bool) bool {
	if elevated {
		return elevated
	}
	return os.Geteuid() == 0
}

func isVirtualDisk(dsk disk.PartitionStat) bool {
	// TODO: find a way to detect virtual disks? /mnt/share or /Volumes/Share ?
	// - Linux: Parse /etc/mtab or /proc/mounts ?
	// - macOS: Use the mount command and parse output ?
	slog.Error("isVirtualDisk() NOT implemented for Linux!")
	return false
}
