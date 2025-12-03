package gtm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
)

const (
	GIBIBYTE = 1_073_741_824 // 2^30 - binary gigabyte
	GIGABYTE = 1_000_000_000 // 10^9 - decimal gigabyte
)

type FileSystem int

const (
	unknownFS FileSystem = iota
	apfs
	btrfs
	fat
	fat32
	exfat
	ext
	ext2
	ext3
	ext4
	ntfs
	jfs
	ZFS
)

const (
	cpuUpdateInterval      time.Duration = time.Second
	diskUpdateInterval     time.Duration = time.Minute
	gpuUpdateInterval      time.Duration = time.Second
	hostInfoUpdateInterval time.Duration = time.Second
	memUpdateInterval      time.Duration = time.Second
	netUpdateInterval      time.Duration = time.Second
	// procsUpdateInterval    time.Duration = time.Second
)

type CPU struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Vendor        string `json:"vendor"`
	CountPhysical int    `json:"countPhysical"`
	CountLogical  int    `json:"countLogical"`
}

type CPUStat struct {
	UsagePercent float64
}

// CPUTempStat Reference: https://wutils.com/wmi/root/wmi/msacpi_thermalzonetemperature/#properties
type CPUTempStat struct {
	Active bool
	// Temperature levels (in tenths of degrees Kelvin) at which the OS must enable active cooling
	ActiveTripPoint      uint32
	ActiveTripPointCount uint32
	// Temperature (in tenths of degrees Kelvin) to shutdown the system (critical temperature)
	CriticalTripPoint  uint32
	CurrentTemperature uint32 // In tenths of degrees Kelvin
	InstanceName       string
	// Temperature (in tenths of degrees Kelvin) at which CPU throttling is activated (enable
	// passive cooling)
	PassiveTripPoint uint32
	Reserved         uint32
	SamplingPeriod   uint32
	ThermalConstant1 uint32
	ThermalConstant2 uint32
	ThermalStamp     uint32
}

type DiskStat struct {
	Mountpoint    string     `json:"mountpoint"`
	Device        string     `json:"device"`
	FSType        FileSystem `json:"fs"`
	IsVirtualDisk bool       `json:"isVirtualDisk"`
	Free          uint64     `json:"free"`
	Used          uint64     `json:"used"`
	UsedPercent   float64    `json:"usedPercent"`
	Total         uint64     `json:"total"`
}

type GPU struct {
	Name   string `json:"name"`
	Vendor string `json:"vendor"`
}

type GPUStat struct {
	Id          int64   `json:"id"`
	Load        float64 `json:"load"`
	MemoryUsage float64 `json:"memoryUsage"`
	MemoryTotal float64 `json:"memoryTotal"`
	Power       float64 `json:"power"`
	Temperature int64   `json:"temperature"`
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
	// lastFetchProc time.Time.
)

var (
	IsAdmin  bool
	HasGPU   bool
	hostname string
)

func init() {
	gpuInfo = &GPU{}
	// Passing in the IsAdmin var is useful to avoid making duplicate syscalls if checkIsAdmin()
	// 	has ever returned true.
	IsAdmin = isAdmin(IsAdmin)
	HasGPU = hasGPU()
}

func ConvertBytesToGB(bytes uint64, rounded bool) float64 {
	result := float64(bytes) / GIGABYTE
	if rounded {
		return math.RoundToEven(result)
	}

	return result
}

func ConvertBytesToGiB(bytes uint64, rounded bool) float64 {
	result := float64(bytes) / GIBIBYTE
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

	info, err := cpu.Info()
	if err != nil {
		slog.Error("Failed to retrieve cpu.Info()! " + err.Error())
	}

	for _, c := range info {
		slog.Debug("cpu.Info(): "+c.String(), "socketCount", len(info))
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

func CPUModelName() string {
	if cpuInfo == nil {
		CPUInfo()
	}

	replacer := strings.NewReplacer(
		"(R)", "",
		"(TM)", "",
		"CPU @ ", "@",
		"Core ", "",
	)

	var cpuName string
	if cpuInfo[0].Vendor == "GenuineIntel" {
		cpuName = replacer.Replace(cpuInfo[0].Name)
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
	if len(cpuStats) > 0 && time.Since(lastFetchCPU) < cpuUpdateInterval {
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
	temps, _ := sensors.SensorsTemperatures()

	for _, temp := range temps {
		if len(cpuTemp) > 0 && temp.Temperature != cpuTemp[len(cpuTemp)-1] {
			return "ok", nil
		}
	}
	// slog.Debug("GetCPUTemp(): " + fmt.Sprintf("%+v", temps))
	return fmt.Sprintf("%v", temps), nil
}

func convertFSTypeToEnum(fsType string) FileSystem {
	switch fsType {
	case "APFS":
		return apfs
	case "BTRFS":
		return btrfs
	case "FAT":
		return fat
	case "FAT32":
		return fat32
	case "exFAT":
		return exfat
	case "EXT":
		return ext
	case "EXT2":
		return ext2
	case "EXT3":
		return ext3
	case "EXT4":
		return ext4
	case "NTFS":
		return ntfs
	case "JFS":
		return jfs
	case "ZFS":
		return ZFS
	default:
		return unknownFS // unknown filesystem?
	}
}

func DisksStats() []DiskStat {
	if time.Since(lastFetchDisk) < diskUpdateInterval && len(disksStats) > 0 {
		return disksStats
	}

	disks, err := disk.Partitions(Cfg.DetectVirtualDrives)
	if err != nil {
		slog.Error("Failed to retrieve disk.Partitions()! " + err.Error())
	}

	lastFetchDisk = time.Now()

	disksStats = make([]DiskStat, len(disks))
	for i, dsk := range disks {
		slog.Debug(">> DISK " + dsk.String())

		usage, err := disk.Usage(dsk.Mountpoint)
		if err != nil {
			slog.Error("Failed to retrieve disk.Usage(" + dsk.Mountpoint + ")! " + err.Error())
		}

		slog.Debug("usage: " + usage.String())

		// If the drive is a virtual disk and DetectVirtualDrives is true, then track them.
		// Otherwise, only report back physical disks (non-virtual disks)
		isVDisk := isVirtualDisk(dsk)
		if Cfg.DetectVirtualDrives && isVDisk || !Cfg.DetectVirtualDrives && !isVDisk {
			disksStats[i] = DiskStat{
				Mountpoint:    dsk.Mountpoint,
				Device:        dsk.Device,
				FSType:        convertFSTypeToEnum(usage.Fstype),
				IsVirtualDisk: isVDisk,
				Free:          usage.Free,
				Used:          usage.Used,
				UsedPercent:   math.Round((usage.UsedPercent*100)/100) / 100,
				Total:         usage.Total,
			}
		}
	}

	return disksStats
}

func hasGPU() bool {
	// If hasGPU is true, then we have checked already. Just return the true
	//  value instead of calling GPU utils again
	if HasGPU {
		return HasGPU
	}

	if err := exec.CommandContext(context.Background(), "nvidia-smi").Run(); err == nil {
		gpuInfo.Vendor = "nvidia"
		HasGPU = true

		return HasGPU
	}

	if err := exec.CommandContext(context.Background(), "rocm-smi").Run(); err == nil {
		gpuInfo.Vendor = "amd"
		HasGPU = true

		return HasGPU
	}

	slog.Error("HasGPU(): Could not find NVIDIA or AMD GPUs installed using SMI")

	return HasGPU
}

func (g *GPUStat) String() string {
	// NVIDIA always reports memory usage in MiB
	memoryUsageGiB := fmt.Sprintf("%.0f", g.MemoryUsage) ///1024)
	memoryTotalGiB := fmt.Sprintf("%.0f", g.MemoryTotal) ///1024)

	// memoryUsageGiB = fmt.Sprintf("%.2f", (g.MemoryUsage/g.MemoryTotal))

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

	info := strings.SplitSeq(string(output), "\n")
	for line := range info {
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
				Id:          id,
				Load:        float64(load) / 100,
				MemoryUsage: memoryUsage,
				MemoryTotal: memoryTotal,
				Power:       power,
				Temperature: temp,
			}
			gpuStats = append(gpuStats, gpu)
		}
	}

	return gpuStats
}

func GPUStats() []GPUStat {
	// Limit getting device data to just once a second, and NOT with every UI update
	if time.Since(lastFetchGPU) < gpuUpdateInterval && gpuStats != nil {
		return gpuStats
	}

	switch gpuInfo.Vendor {
	case "nvidia":
		cmd := exec.CommandContext(context.Background(),
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
		// slog.Debug(data[len(data)-1].String())
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
	if time.Since(lastFetchHost) < hostInfoUpdateInterval && len(hostInfo.String()) > 0 {
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
	if time.Since(lastFetchMem) < memUpdateInterval && len(memStats.String()) > 0 {
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
		// slog.Debug("gtm.MemoryStats(): no changes... return last fetch")
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
	if time.Since(lastFetchNet) < netUpdateInterval && len(netInfo) > 0 {
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
