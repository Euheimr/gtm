package gtm

import (
	"log/slog"
	"strconv"

	"github.com/shirou/gopsutil/v4/disk"
	"golang.org/x/sys/windows"
)

// Returns true if the current process is running as admin or elevated privileges. Primarily
// useful for checking system temperature sensors. The 'elevated' parameter is useful for
// avoiding duplicate syscalls if we've already checked isAdmin() and it returned true previously.
func isAdmin(elevated bool) bool {
	if elevated {
		return elevated
	}

	isElevated := windows.GetCurrentProcessToken().IsElevated()
	slog.Debug("isAdmin(): " + strconv.FormatBool(isElevated))

	return isElevated
}

// Virtual disks are non-physical networked drives; NAS / Google Drive / Dropbox / etc.
func isVirtualDisk(dsk disk.PartitionStat) bool {
	dskPath, err := windows.UTF16PtrFromString(dsk.Mountpoint)
	if err != nil {
		slog.Error("isVirtualDisk(): Failed to get UTF16 pointer from string - " +
			dsk.Mountpoint + "! " + err.Error())
	}

	driveType := windows.GetDriveType(dskPath)

	// 2: DRIVE_REMOVABLE 3: DRIVE_FIXED 4: DRIVE_REMOTE 5: DRIVE_CDROM 6: DRIVE_RAMDISK
	switch driveType {
	case windows.DRIVE_FIXED:
		// disk.IOCounters(C:) ALWAYS errors out on Windows, HOWEVER, we do not get an
		//	empty struct on a valid DRIVE_FIXED device
		io, _ := disk.IOCounters(dsk.Mountpoint)
		slog.Debug("disk.IOCounters(" + dsk.Mountpoint + "): " + io[dsk.Mountpoint].String())
		slog.Debug("isVirtualDisk(): " + dsk.Mountpoint + " is a FIXED drive")

		return false
	case windows.DRIVE_CDROM:
		slog.Debug("isVirtualDisk(): " + dsk.Mountpoint + " is a CDROM drive")

		return false
	case windows.DRIVE_REMOVABLE:
		slog.Debug("isVirtualDisk(): " + dsk.Mountpoint + " is a REMOVABLE drive")

		return false
	case windows.DRIVE_RAMDISK:
		slog.Debug("isVirtualDisk(): " + dsk.Mountpoint + " is a RAMDISK")

		return true
	case windows.DRIVE_REMOTE:
		slog.Debug("isVirtualDisk(): " + dsk.Mountpoint + " is a REMOTE drive")

		return true
	default:
		slog.Debug("isVirtualDisk(): " + dsk.Mountpoint + " is an UNKNOWN drive type")

		return false
	}
}
