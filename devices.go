package gtm

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
	"log/slog"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

const GIBIBYTE = 1_073_741_824 // binary base 2^30 or 1024^3
//const GIGABYTE = 1_000_000_000 // base 10^9

type GPUData struct {
	Id          int32   `json:"id"`
	Load        float64 `json:"load"`
	MemoryUsage float64 `json:"memoryUsage"`
	MemoryTotal float64 `json:"memoryTotal"`
	Power       float64 `json:"power"`
	Temperature int32   `json:"temperature"`
}

var (
	cpuInfo  []cpu.InfoStat
	diskInfo []disk.PartitionStat
	//gpuInfo  *[]GPUData
	hostInfo *host.InfoStat
	memInfo  *mem.VirtualMemoryStat
	netInfo  []net.IOCountersStat
	sensInfo []sensors.TemperatureStat
)

var (
	cpuModelName string
	hostname     string
	gpuName      string
)

var (
	GPUVendor string
)

func init() {}

func ConvertBytesToGiB(bytes uint64, rounded bool) (result float64) {
	result = float64(bytes) / GIBIBYTE
	if rounded {
		// effectively return an integer via rounding the float to an int (ie. "11.0" GB)
		return math.RoundToEven(result)
	} else {
		return result
	}
}

func GetCPUInfo() []cpu.InfoStat {
	cInfo, err := cpu.Info()
	if err != nil {
		slog.Error("Failed to retrieve cpu.Info()! " + err.Error())
	}
	cpuInfo = cInfo
	slog.Debug("cpu.Info(): "+cpuInfo[0].String(), "socketCount", len(cpuInfo))

	// model name doesn't change with each syscall... so cache it here
	cpuModelName = cpuInfo[0].ModelName

	return cpuInfo
}

func GetCPUModel(formatName bool) string {
	if cpuModelName == "" {
		GetCPUInfo()
	}
	cpuModel := cpuModelName
	if formatName && cpuInfo[0].VendorID == "GenuineIntel" {
		cpuModel = strings.ReplaceAll(cpuModel, "(R)", "")
		cpuModel = strings.ReplaceAll(cpuModel, "(TM)", "")
		cpuModel = strings.ReplaceAll(cpuModel, "CPU @ ", "@")
		cpuModel = strings.ReplaceAll(cpuModel, "Core ", "")
	}
	// TODO: format AMD & ARM ?
	return cpuModel
}

func GetDiskInfo() []disk.PartitionStat {
	dInfo, err := disk.Partitions(false)
	if err != nil {
		slog.Error("Failed to retrieve disk.Partitions()! " + err.Error())
	}
	diskInfo = dInfo
	slog.Debug("disk.Partitions(): physical disk count = " + strconv.Itoa(len(diskInfo)))
	for i, dsk := range diskInfo {
		slog.Debug("disk.Partitions(): disk #" + strconv.Itoa(i) + ": " + dsk.String())
	}
	return diskInfo
}

func parseGPUNvidiaData(output []byte) []GPUData {
	var gpuData []GPUData

	info := strings.Split(string(output), "\n")
	for _, line := range info {
		if line != "" {
			data := strings.Split(line, ", ")
			gpuName = data[1]

			id, err := strconv.ParseInt(data[0], 10, 32)
			if err != nil {
				slog.Error("Failed to parse GPU Id from string -> int ! " + err.Error())
			}

			load, err := strconv.ParseInt(data[2], 10, 32)
			if err != nil {
				slog.Error("Failed to parse GPU Load from string -> int ! " + err.Error())
			}

			memoryUsage, err := strconv.ParseFloat(data[3], 64)
			if err != nil {
				slog.Error("Failed to parse float: memory.usage !" + err.Error())
				memoryUsage = 0.0
			}
			memoryTotal, err := strconv.ParseFloat(data[4], 64)
			if err != nil {
				slog.Error("Failed to parse float: memory.total !" + err.Error())
				memoryTotal = 0.0
			}

			power, err := strconv.ParseFloat(data[5], 64)
			if err != nil {
				slog.Error("Failed to parse float: power !" + err.Error())
			}

			t := strings.ReplaceAll(data[6], "\r", "")
			temp, err := strconv.ParseInt(t, 10, 32)
			if err != nil {
				slog.Error("Failed to parse float: temp !" + err.Error())
			}

			gpu := GPUData{
				Id:          int32(id),
				Load:        float64(load) / 100,
				MemoryUsage: memoryUsage,
				MemoryTotal: memoryTotal,
				Power:       power,
				// on windows, there's a carriage return on the last stat
				Temperature: int32(temp),
			}
			gpuData = append(gpuData, gpu)
		}
	}
	return gpuData
}

func getGPUNvidiaData() ([]GPUData, error) {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=index,name,utilization.gpu,memory.used,memory.total,power.draw,temperature.gpu",
		"--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		slog.Error("Failed to get nvidia-smi output! " + err.Error())
		return nil, err
	}
	return parseGPUNvidiaData(output), nil
}

func HasGPU() bool {
	if err := exec.Command("nvidia-smi").Run(); err == nil {
		GPUVendor = "nvidia"
		return true
	}
	if err := exec.Command("rocm-smi").Run(); err == nil {
		GPUVendor = "amd"
		return true
	}
	slog.Error("HasGPU(): Could not find NVIDIA or AMD GPUs installed using SMI")
	return false
}

func (g *GPUData) String() string {
	// NVIDIA always reports memory usage in MiB
	memoryUsageGiB := fmt.Sprintf("%.0f", g.MemoryUsage) ///1024)
	memoryTotalGiB := fmt.Sprintf("%.0f", g.MemoryTotal) ///1024)

	//memoryUsageGiB = fmt.Sprintf("%.2f", (g.MemoryUsage/g.MemoryTotal))

	return fmt.Sprintf("gfx card #%v, %v%%, %v MiB, %v MiB, %vW, %v°C",
		g.Id, int(g.Load*100), memoryUsageGiB, memoryTotalGiB, g.Power, g.Temperature)
}

func (g *GPUData) JSON(indent bool) string {
	if indent {
		out, err := json.MarshalIndent(g, "", "  ")
		if err != nil {
			slog.Error("Failed to marshal JSON from struct GPUData{} ! " + err.Error())
			return ""
		}
		return string(out)
	} else {
		out, err := json.Marshal(g)
		if err != nil {
			slog.Error("Failed to marshal JSON from struct GPUData{} ! " + err.Error())
		}
		return string(out)
	}
}

func GetGPUInfo() []GPUData {
	if HasGPU {
		switch GPUVendor {
		case "nvidia":
			data, err := getGPUNvidiaData()
			if err != nil {
				slog.Error("Failed to retrieve NVIDIA GPU data from nvidia-smi ! " + err.Error())
			}
			lastDataStat := data[len(data)-1]
			slog.Debug(lastDataStat.String())
			return data

		case "amd":
			// TODO: write rocm-smi code for AMD gpu detection and data parsing
			slog.Error("AMD GPU not implemented yet !")
			return nil
		}
	}
	return nil
}

func GetGPUName() string { return gpuName }

func GetHostInfo() *host.InfoStat {
	hInfo, err := host.Info()
	if err != nil {
		slog.Error("Failed to retrieve host.Info()! " + err.Error())
	}
	hostInfo = hInfo
	slog.Debug("host.Info(): " + hostInfo.String())

	hostname = hostInfo.Hostname
	return hostInfo
}

func GetHostname() string {
	if hostname != "" {
		return hostname
	} else {
		GetHostInfo()
		return hostname
	}
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
