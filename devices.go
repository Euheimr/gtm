package gtm

import (
	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/gpu"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"log/slog"
	"strings"
)

var hostname string

var gpuInfo *gpu.Info

func init() {
	//cInfo, err := cpu.Info()
	//if err != nil {
	//	slog.Error("Failed to retrieve cpu.Info()! " + err.Error())
	//}
	//cpuInfo = cInfo

	//dInfo, err := disk.Usage("/")
	//if err != nil {
	//	slog.Error("Failed to retrieve disk.Usage()! " + err.Error())
	//}
	//diskInfo = dInfo

	//mInfo, err := mem.VirtualMemory()
	//if err != nil {
	//	slog.Error("Failed to retrieve mem.VirtualMemory()! " + err.Error())
	//}
	//memInfo = mInfo

	//nInfo, err := net.IOCounters(true)
	//if err != nil {
	//	slog.Error("Failed to retrieve net.IOCounters()! " + err.Error())
	//}
	//netInfo = &nInfo[0]

}

func GetHostInfo() *host.InfoStat {
	hostInfo, err := host.Info()
	if err != nil {
		slog.Error("Failed to retrieve host.Info()! " + err.Error())
	}
	hostname = hostInfo.Hostname
	slog.Debug("host.Info(): " + hostInfo.String())
	return hostInfo
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

func GetCpuModel() string {
	cpuInfo, err := cpu.Info()
	if err != nil {
		slog.Error("Failed to retrieve cpu.Info()! " + err.Error())
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
	if hostname == "" {
		hostInfo, err := host.Info()
		if err != nil {
			slog.Error("Failed to retrieve host.Info()! " + err.Error())
		}
		hostname = hostInfo.Hostname
	}
	return hostname
}
