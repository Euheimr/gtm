package gtm

import (
	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/gpu"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
	"log/slog"
	"math"
	"strconv"
	"strings"
)

const GIGABYTE = 1_073_741_824

var (
	cpuInfo  []cpu.InfoStat
	diskInfo []disk.PartitionStat
	gpuInfo  *gpu.Info
	hostInfo *host.InfoStat
	memInfo  *mem.VirtualMemoryStat
	netInfo  []net.IOCountersStat
	sensInfo []sensors.TemperatureStat
)

func init() {

	//dInfo, err := disk.Usage("/")
	//if err != nil {
	//	slog.Error("Failed to retrieve disk.Usage()! " + err.Error())
	//}
	//diskInfo = dInfo
}

func ConvertBytesToGB(ramBytes uint64, round bool) (result float64) {
	result = float64(ramBytes) / GIGABYTE
	if !round {
		return result
	} else {
		// effectively return an integer via rounding the float (ie. "11.0" GB)
		return math.RoundToEven(result)
	}
}

func GetCPUInfo() []cpu.InfoStat {
	cInfo, err := cpu.Info()
	if err != nil {
		slog.Error("Failed to retrieve cpu.Info()! " + err.Error())
	}
	cpuInfo = cInfo
	slog.Debug("cpu.Info(): "+cpuInfo[0].String(), "socketCount", len(cpuInfo))
	return cpuInfo
}

func GetDiskInfo() {
	dInfo, err := disk.Partitions(false)
	if err != nil {
		slog.Error("Failed to retrieve disk.Partitions()! " + err.Error())
	}
	diskInfo = dInfo
	slog.Debug("disk.Partitions(): physical disk count = " + strconv.Itoa(len(diskInfo)))
	for i, dsk := range diskInfo {
		slog.Debug("disk.Partitions(): disk #" + strconv.Itoa(i) + ": " + dsk.String())
	}
}

func GetGPUInfo() *gpu.Info {
	gInfo, err := ghw.GPU()
	if err != nil {
		slog.Error("Failed to retrieve gpu.Info()! " + err.Error())
		return nil
	}
	gpuInfo = gInfo
	slog.Debug("gpu.Info(): " + gpuInfo.String())
	return gpuInfo
}

func GetHostInfo() *host.InfoStat {
	hInfo, err := host.Info()
	if err != nil {
		slog.Error("Failed to retrieve host.Info()! " + err.Error())
	}
	hostInfo = hInfo
	slog.Debug("host.Info(): " + hostInfo.String())
	return hostInfo
}

func GetMemoryInfo() *mem.VirtualMemoryStat {
	mInfo, err := mem.VirtualMemory()
	if err != nil {
		slog.Error("Failed to retrieve mem.VirtualMemory()! " + err.Error())
	}

	if memInfo == nil {
		// If this is the first time getting the memory usage, just populate/init memInfo
		memInfo = mInfo
		return memInfo
	} else {
		currentUsedPercent := mInfo.UsedPercent
		oldUsedPercent := memInfo.UsedPercent

		if currentUsedPercent == oldUsedPercent {
			// If we get the same results, just re-send the same data without updates
			//slog.Debug("gtm.GetMemoryInfo(): no changes... return last fetch")
			return memInfo
		} else {
			//  If the previous fetch is greater than or less than the last fetch in
			// 	Gigabytes, return the updated memory usage
			memInfo = mInfo
			slog.Debug("mem.VirtualMemory(): " + memInfo.String())
			return memInfo
		}
	}
}

func GetNetworkInfo() []net.IOCountersStat {
	nInfo, err := net.IOCounters(false)
	if err != nil {
		slog.Error("Failed to retrieve net.IOCounters()! " + err.Error())
	}
	netInfo = nInfo
	for i, iface := range netInfo {
		slog.Debug("net.IOCounters(), interface #" + strconv.Itoa(i) + ": " +
			iface.String())
	}
	return netInfo
}

func GetSensorsInfo() []sensors.TemperatureStat {
	sInfo, err := sensors.SensorsTemperatures()
	if err != nil {
		slog.Error("Failed to retrieve sensors.SensorsTemperatures()! " + err.Error())
	}
	sensInfo = sInfo
	for i, sensor := range sensInfo {
		slog.Debug("sensors.SensorsTemperatures(), sensor #" + strconv.Itoa(i) + ": " +
			sensor.String())
	}
	return sensInfo
}

func GetCpuModel() string {
	if cpuInfo == nil {
		GetCPUInfo()
	}
	cpuModel := cpuInfo[0].ModelName
	if cpuInfo[0].VendorID == "GenuineIntel" {
		cpuModel = strings.ReplaceAll(cpuModel, "(R)", "")
		cpuModel = strings.ReplaceAll(cpuModel, "(TM)", "")
		cpuModel = strings.ReplaceAll(cpuModel, "CPU @ ", "@")
		cpuModel = strings.ReplaceAll(cpuModel, "Core ", "")
	}

	return cpuModel
}

func HasGPU() bool {
	if gpuInfo == nil {
		return false
	} else if Cfg.EnableGPU && len(gpuInfo.GraphicsCards) > 0 {
		return true
	} else {
		return false
	}
}

func GetGPUName() string {
	if HasGPU() {
		return gpuInfo.GraphicsCards[0].DeviceInfo.Product.Name
	} else {
		return "NO GPU"
	}
}

func GetHostname() string {
	if hostInfo == nil {
		GetHostInfo()
	}
	return hostInfo.Hostname
}
