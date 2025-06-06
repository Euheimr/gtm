//go:build windows

package gtm

import (
	"github.com/shirou/gopsutil/v4/disk"
	"golang.org/x/sys/windows"
	"log/slog"
	"runtime"
)

func isVirtualDisk(path string) bool {
	switch runtime.GOOS {
	case "windows":
		d, err := windows.UTF16PtrFromString(path)
		if err != nil {
			slog.Error("Failed to get UTF16 pointer from string: " + path + "! " +
				err.Error())
		}
		driveType := windows.GetDriveType(d)

		// 2: DRIVE_REMOVABLE 3: DRIVE_FIXED 4: DRIVE_REMOTE 5: DRIVE_CDROM 6: DRIVE_RAMDISK
		switch driveType {
		case windows.DRIVE_RAMDISK:
			slog.Debug(path + " is a RAMDISK")
			return true
		case windows.DRIVE_FIXED:
			// disk.IOCounters(C:) ALWAYS errors out on Windows, HOWEVER, we do not get an
			//	empty struct on a valid DRIVE_FIXED device
			io, _ := disk.IOCounters(path)
			switch len(io) {
			case 0:
				// This is a VERY hacky way of working around detecting Google Drive.
				//	GDrive is seen as a "real" drive in Windows for some reason, and
				//	not as a RAMDISK (Virtual Hard Disk; aka. VHD).
				// But if we try to call disk.IOCounters() on it, we will just get an
				//	empty struct (length of 0) back, which indicates it IS a RAMDISK.
				// This is the only way I've been able to detect a mounted Google
				//	Drive :(
				slog.Debug("drive " + path + " IS a RAMDISK")
				return true
			default:
				// Any other case that is len(io) > 0 means it is not a RAMDISK
				slog.Debug("disk.IOCounters(" + path + "): " + io[path].String())
				return false
			}
		default:
			slog.Debug(path + " is not a RAMDISK")
			return false
		}
	default:
		// TODO: do RAMDISK checks for macOS & Linux !
		slog.Debug("Not on windows... ignoring RAMDISK check ...")
		return false
	}
}
