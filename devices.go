package gtm

import (
	"encoding/json"
	"fmt"
	"github.com/euheimr/ringbuffer"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
	"golang.org/x/sys/windows"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// GIBIBYTE is the binary representation of gigabyte
const GIBIBYTE = 1_073_741_824 // Binary base 2^30 or 1024^3
const GIGABYTE = 1_000_000_000 // Decimal base 10^9

type FileSystemType int

const (
	APFS FileSystemType = iota
	exFAT
	FAT
	FAT32
	EXT
	EXT2
	EXT3
	EXT4
	NTFS
	JFS
	ZFS
)

const (
	CPU_STATS_UPDATE_INTERVAL  = time.Second
	DISK_STATS_UPDATE_INTERVAL = time.Minute
	GPU_STATS_UPDATE_INTERVAL  = time.Second
	HOST_INFO_UPDATE_INTERVAL  = time.Second
	MEM_STATS_UPDATE_INTERVAL  = time.Second
	NET_STATS_UPDATE_INTERVAL  = time.Second
	PROCS_UPDATE_INTERVAL      = time.Second
)

type CPU struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Vendor        string `json:"vendor"`
	CountPhysical int    `json:"count_physical"`
	CountLogical  int    `json:"count_logical"`
}

type CPUStat struct {
	UsagePercent float64
}

// CPUTempStat
// Reference: https://wutils.com/wmi/root/wmi/msacpi_thermalzonetemperature/#properties
type CPUTempStat struct {
	Active               bool
	ActiveTripPoint      uint32 // Temperature levels (in tenths of degrees Kelvin) at which the OS must activate active cooling
	ActiveTripPointCount uint32 // Count of active trip points
	CriticalTripPoint    uint32 // Temperature (in tenths of degrees Kelvin) at which the OS must shutdown the system (ie, critical temperature)
	CurrentTemperature   uint32 // Temperature at thermal zone in tenths of degrees Kelvin
	InstanceName         string // ??
	PassiveTripPoint     uint32 // Temperature (in tenths of degrees Kelvin) at which the OS must activate CPU throttling (ie, enable passive cooling)
	Reserved             uint32 // ??
	SamplingPeriod       uint32 // ??
	ThermalConstant1     uint32 // First thermal constant
	ThermalConstant2     uint32 // Second thermal constant
	ThermalStamp         uint32 // Thermal information change stamp
}

type DiskStat struct {
	Mountpoint    string         `json:"mountpoint"`
	Device        string         `json:"device"`
	FSType        FileSystemType `json:"fs_type"`
	IsVirtualDisk bool           `json:"is_virtual_disk"`
	Free          uint64         `json:"free"`
	Used          uint64         `json:"used"`
	UsedPercent   float64        `json:"used_percent"`
	Total         uint64         `json:"total"`
}

type GPU struct {
	Name   string
	Vendor string
}

type GPUStat struct {
	Id          int32   `json:"card-id"`
	Load        float64 `json:"load"`
	MemoryUsage float64 `json:"memoryUsage"`
	MemoryTotal float64 `json:"memoryTotal"`
	Power       float64 `json:"power"`
	Temperature int32   `json:"temperature"`
}

type GPURingBuffer struct {
	Load        *ringbuffer.RingBuffer[float32]
	MemoryUsage *ringbuffer.RingBuffer[float32]
	MemoryTotal *ringbuffer.RingBuffer[float32]
	Power       *ringbuffer.RingBuffer[float32]
	Temperature *ringbuffer.RingBuffer[float32]
}

var (
	cpuInfo    []CPU
	cpuStats   []CPUStat
	cpuTemp    []float64
	disksStats []DiskStat
	gpuInfo    *GPU
	gpuStats   []GPUStat
	hostInfo   *host.InfoStat
	memStats   *mem.VirtualMemoryStat
	netInfo    []net.IOCountersStat
)

var (
	lastFetchCPU  time.Time
	lastFetchDisk time.Time
	lastFetchGPU  time.Time
	lastFetchHost time.Time
	lastFetchMem  time.Time
	lastFetchNet  time.Time
	lastFetchProc time.Time
)

var (
	isAdmin  bool
	hasGPU   bool
	hostname string
)

func init() {
	gpuInfo = &GPU{}
	isAdmin = checkIsAdmin()
}

func checkIsAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	return true
}

func ConvertBytesToGB(bytes uint64, rounded bool) (result float64) {
	result = float64(bytes) / GIGABYTE
	if rounded {
		return math.RoundToEven(result)
	}
	return result
}

func ConvertBytesToGiB(bytes uint64, rounded bool) (result float64) {
	result = float64(bytes) / GIBIBYTE
	if rounded {
		// Effectively return an integer via rounding the float to an int
		//	(ie. "11.0" GiB)
		return math.RoundToEven(result)
	}
	return result
}

func CPUInfo() []CPU {
	if cpuInfo != nil {
		return cpuInfo
	}

	cInfo, err := cpu.Info()
	if err != nil {
		slog.Error("Failed to retrieve cpu.Info()! " + err.Error())
	}
	for _, c := range cInfo {
		slog.Debug("cpu.Info(): "+c.String(), "socketCount", len(cInfo))
		info := &CPU{
			// model name doesn't change with each syscall... so cache it here
			Name:          c.ModelName,
			Vendor:        c.VendorID,
			CountPhysical: int(c.Cores),
			CountLogical:  0,
		}
		cpuInfo = append(cpuInfo, *info)
	}
	return cpuInfo
}

func CPUModelName() (cpuName string) {
	if cpuInfo == nil {
		CPUInfo()
	}
	if cpuInfo[0].Vendor == "GenuineIntel" {
		cpuName = strings.ReplaceAll(cpuInfo[0].Name, "(R)", "")
		cpuName = strings.ReplaceAll(cpuInfo[0].Name, "(TM)", "")
		cpuName = strings.ReplaceAll(cpuInfo[0].Name, "CPU @ ", "@")
		cpuName = strings.ReplaceAll(cpuInfo[0].Name, "Core ", "")
	}
	// TODO: format AMD & ARM ?

	return cpuName
}

func (c CPU) String() string {
	return fmt.Sprintf(
		"socket #%v, name=%s, vendor=%s, countPhys=%v, countLogical=%v",
		c.Id, c.Name, c.Vendor, c.CountPhysical, c.CountLogical)
}

func (c CPU) JSON(indent bool) string {
	if indent {
		out, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			slog.Error("Failed to marshal indent JSON from struct CPU{} !" + err.Error())
		}
		return string(out)
	} else {
		out, err := json.Marshal(c)
		if err != nil {
			slog.Error("Failed to marshal JSON from struct CPU{} !" + err.Error())
		}
		return string(out)
	}
}

func GetCPUStats() []CPUStat {
	if len(cpuStats) > 0 && time.Since(lastFetchCPU) < CPU_STATS_UPDATE_INTERVAL {
		return cpuStats
	}
	cpuPct, err := cpu.Percent(0, false)
	if err != nil {
		slog.Error("Failed to fetch cpu.Percent() !" + err.Error())
	}
	lastFetchCPU = time.Now()

	stats := CPUStat{
		UsagePercent: cpuPct[0],
	}
	// TODO: fetch cpu usage and append to data
	cpuStats = append(cpuStats, stats)

	return cpuStats
}

func CPUTemp() (string, error) {
	if !isAdmin {
		return windows.ERROR_ACCESS_DENIED.Error(), windows.ERROR_ACCESS_DENIED
	}
	temps, _ := sensors.SensorsTemperatures()

	for _, temp := range temps {
		if len(cpuTemp) > 0 && temp.Temperature != cpuTemp[len(cpuTemp)-1] {
			return fmt.Sprint("ok"), nil
		}
	}
	slog.Debug("GetCPUTemp(): " + fmt.Sprintf("%+v", temps))
	return fmt.Sprintf("%v", temps), nil
}

func convertFSType(fsType string) FileSystemType {
	switch fsType {
	case "APFS":
		return APFS
	case "exFAT":
		return exFAT
	case "FAT":
		return FAT
	case "FAT32":
		return FAT32
	case "EXT":
		return EXT
	case "EXT2":
		return EXT2
	case "EXT3":
		return EXT3
	case "EXT4":
		return EXT4
	case "NTFS":
		return NTFS
	case "JFS":
		return JFS
	case "ZFS":
		return ZFS
	default:
		// unknown filesystem?
		return -1
	}
}

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

func DisksStats() []DiskStat {
	if time.Since(lastFetchDisk) < DISK_STATS_UPDATE_INTERVAL && len(disksStats) > 0 {
		return disksStats
	}

	dInfo, err := disk.Partitions(false)
	if err != nil {
		slog.Error("Failed to retrieve disk.Partitions()! " + err.Error())
	}
	lastFetchDisk = time.Now()

	disksStats = make([]DiskStat, len(dInfo))
	for i, dsk := range dInfo {
		usage, err := disk.Usage(dsk.Mountpoint)
		if err != nil {
			slog.Error("Failed to retrieve disk.Usage(" + dsk.Mountpoint + ")! " +
				err.Error())
		}
		slog.Debug("disk: " + dsk.String())
		slog.Debug("usage: " + usage.String())

		// convert filesystem type to integer
		fsType := convertFSType(usage.Fstype)
		isVDisk := isVirtualDisk(dsk.Mountpoint)
		usedPercent := math.Round((usage.UsedPercent*100)/100) / 100

		stats := DiskStat{
			Mountpoint:    dsk.Mountpoint,
			Device:        dsk.Device,
			FSType:        fsType,
			IsVirtualDisk: isVDisk,
			Free:          usage.Free,
			Used:          usage.Used,
			UsedPercent:   usedPercent,
			Total:         usage.Total,
		}
		disksStats[i] = stats
	}
	return disksStats
}

func HasGPU() bool {
	// Booleans initialize to false. If hasGPU is true, then we have checked already.
	//	Just return the true value instead of calling GPU utils again
	if hasGPU {
		return hasGPU
	}
	if err := exec.Command("nvidia-smi").Run(); err == nil {
		gpuInfo.Vendor = "nvidia"
		hasGPU = true
		return hasGPU
	}
	if err := exec.Command("rocm-smi").Run(); err == nil {
		gpuInfo.Vendor = "amd"
		hasGPU = true
		return hasGPU
	}
	slog.Error("HasGPU(): Could not find NVIDIA or AMD GPUs installed using SMI")
	return hasGPU
}

func (g *GPUStat) String() string {
	// NVIDIA always reports memory usage in MiB
	memoryUsageGiB := fmt.Sprintf("%.0f", g.MemoryUsage) ///1024)
	memoryTotalGiB := fmt.Sprintf("%.0f", g.MemoryTotal) ///1024)

	//memoryUsageGiB = fmt.Sprintf("%.2f", (g.MemoryUsage/g.MemoryTotal))

	return fmt.Sprintf("gfx card #%v, %v%%, %v MiB, %v MiB, %vW, %v°C",
		g.Id, int(g.Load*100), memoryUsageGiB, memoryTotalGiB, g.Power, g.Temperature)
}

func (g *GPUStat) JSON(indent bool) string {
	if indent {
		out, err := json.MarshalIndent(g, "", "  ")
		if err != nil {
			slog.Error("Failed to marshal indent JSON from struct GPUStat{} ! " +
				err.Error())
		}
		return string(out)
	} else {
		out, err := json.Marshal(g)
		if err != nil {
			slog.Error("Failed to marshal JSON from struct GPUStat{} ! " + err.Error())
		}
		return string(out)
	}
}

func parseGPUNvidiaStats(output []byte) []GPUStat {
	var (
		id          int64
		load        int64
		memoryUsage float64
		memoryTotal float64
		power       float64
		temp        int64
		err         error
	)

	info := strings.Split(string(output), "\n")
	for _, line := range info {
		if line != "" {
			data := strings.Split(line, ", ")
			gpuInfo.Name = data[1]

			if id, err = strconv.ParseInt(data[0], 10, 32); err != nil {
				slog.Error("Failed to parse GPU Id from string -> int ! " + err.Error())
			}
			if load, err = strconv.ParseInt(data[2], 10, 32); err != nil {
				slog.Error("Failed to parse GPU Load from string -> int ! " + err.Error())
			}
			if memoryUsage, err = strconv.ParseFloat(data[3], 64); err != nil {
				slog.Error("Failed to parse float: memory.usage !" + err.Error())
				memoryUsage = 0.0
			}
			if memoryTotal, err = strconv.ParseFloat(data[4], 64); err != nil {
				slog.Error("Failed to parse float: memory.total !" + err.Error())
				memoryTotal = 0.0
			}
			if power, err = strconv.ParseFloat(data[5], 64); err != nil {
				slog.Error("Failed to parse float: power !" + err.Error())
			}

			// on windows, there's a carriage return on the last stat
			t := strings.ReplaceAll(data[6], "\r", "")
			if temp, err = strconv.ParseInt(t, 10, 32); err != nil {
				slog.Error("Failed to parse float: temp !" + err.Error())
			}

			gpu := GPUStat{
				Id:          int32(id),
				Load:        float64(load) / 100,
				MemoryUsage: memoryUsage,
				MemoryTotal: memoryTotal,
				Power:       power,
				Temperature: int32(temp),
			}
			gpuStats = append(gpuStats, gpu)
		}
	}
	return gpuStats
}

func GPUStats() []GPUStat {
	// Limit getting device data to just once a second, and NOT with every UI update
	if time.Since(lastFetchGPU) < GPU_STATS_UPDATE_INTERVAL && gpuStats != nil {
		return gpuStats
	}

	switch gpuInfo.Vendor {
	case "nvidia":
		cmd := exec.Command(
			"nvidia-smi",
			"--query-gpu=index,name,utilization.gpu,memory.used,memory.total,"+
				"power.draw,temperature.gpu",
			"--format=csv,noheader,nounits")
		data, err := cmd.Output()
		if err != nil {
			slog.Error("Failed to retrieve NVIDIA GPU data from nvidia-smi ! " +
				err.Error())
			return nil
		}
		//slog.Debug(data[len(data)-1].String())
		gpuStats = parseGPUNvidiaStats(data)
		lastFetchGPU = time.Now()

	case "amd":
		// TODO: write rocm-smi code for AMD gpu detection and data parsing
		slog.Error("AMD GPU not implemented yet !")
		lastFetchGPU = time.Now()
	}
	return gpuStats
}

func GPUName() string { return gpuInfo.Name }

func HostInfo() *host.InfoStat {
	if time.Since(lastFetchHost) < HOST_INFO_UPDATE_INTERVAL && len(hostInfo.String()) > 0 {
		return hostInfo
	}

	hInfo, err := host.Info()
	if err != nil {
		slog.Error("Failed to retrieve host.Info()! " + err.Error())
	}
	lastFetchHost = time.Now()

	hostInfo = hInfo
	slog.Debug("host.Info(): " + hostInfo.String())
	hostname = hostInfo.Hostname

	return hostInfo
}

func Hostname() string {
	if hostname != "" {
		return hostname
	} else {
		HostInfo()
		return hostname
	}
}

func MemoryStats() *mem.VirtualMemoryStat {
	if time.Since(lastFetchMem) < MEM_STATS_UPDATE_INTERVAL && len(memStats.String()) > 0 {
		return memStats
	}

	mStats, err := mem.VirtualMemory()
	if err != nil {
		slog.Error("Failed to retrieve mem.VirtualMemory()! " + err.Error())
	}
	lastFetchMem = time.Now()

	if memStats == nil {
		// This is the first time getting the memory usage; just populate/init memStats
		memStats = mStats
		return memStats
	}

	oldUsedPercent := memStats.UsedPercent
	currentUsedPercent := mStats.UsedPercent

	if oldUsedPercent == currentUsedPercent {
		// If we get the same results, just re-send the same data without updates
		//slog.Debug("gtm.MemoryStats(): no changes... return last fetch")
		return memStats
	} else {
		//  If the previous fetch is greater than or less than the last fetch in
		// 	Gigabytes, return the updated memory usage
		memStats = mStats
		slog.Debug("mem.VirtualMemory(): " + memStats.String())
		return memStats
	}
}

func NetworkStats() []net.IOCountersStat {
	if time.Since(lastFetchNet) < NET_STATS_UPDATE_INTERVAL && len(netInfo) > 0 {
		return netInfo
	}

	nInfo, err := net.IOCounters(false)
	if err != nil {
		slog.Error("Failed to retrieve net.IOCounters()! " + err.Error())
	}
	lastFetchNet = time.Now()

	netInfo = nInfo
	for i, iface := range netInfo {
		slog.Debug("net.IOCounters(), interface #" + strconv.Itoa(i) + ": " +
			iface.String())
	}

	return netInfo
}
