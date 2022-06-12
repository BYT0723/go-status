package main

import (
	"fmt"
	"go-status/util"
	"io/ioutil"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

const (
	leftSplitChar  = "  "
	rightSplitChar = "  "
)

var (
	cpuMute    sync.Mutex
	cpuPercent float64
	netMute    sync.Mutex
	netSpeed   struct {
		DownLoad uint64
		UpLoad   uint64
	}
)

func main() {
	go updateCpu()
	go updateNetSpeed()
	var (
		cmd *exec.Cmd
	)
	for {
		cmd = exec.Command("xsetroot", "-name", getContext())
		cmd.Start()
		time.Sleep(time.Second * 1)
	}

	// for {
	// 	fmt.Printf("getContext(): %v\n", getContext())
	// 	time.Sleep(time.Second * 1)
	// }
}

func getContext() string {
	var primary, extraLeft, extraRight []string

	primary = append(primary, fmt.Sprintf("%2.0f%% %s", cpuPercent, getCpuTemp("%2d°C")))
	primary = append(primary, getMem("%2.1f"))
	primary = append(primary, getDisk("%.1f", "/"))

	extraLeft = append(extraLeft, getNote("#"))

	var (
		format string = "龍%6.1f %s/s %6.1f %s/s"
	)
	down, downUnit := util.CountSize(netSpeed.DownLoad)
	up, upUnit := util.CountSize(netSpeed.UpLoad)
	extraRight = append(extraRight, fmt.Sprintf(format, down, downUnit, up, upUnit))
	extraRight = append(extraRight, getDate("2006-01-02(Mon) 15:04"))

	return fmt.Sprintf("%s;%s;%s",
		rightSplitChar+strings.Join(primary, rightSplitChar),
		strings.Join(extraLeft, leftSplitChar)+leftSplitChar,
		rightSplitChar+strings.Join(extraRight, rightSplitChar))
}

func updateCpu() {
	for {
		percent, _ := cpu.Percent(time.Second, false)
		cpuMute.Lock()
		cpuPercent = percent[0]
		cpuMute.Unlock()
	}
}

func updateNetSpeed() {
	for {
		oldIO, _ := net.IOCounters(false)
		time.Sleep(time.Second * 1)
		newIO, _ := net.IOCounters(false)

		netMute.Lock()
		netSpeed.DownLoad = newIO[0].BytesRecv - oldIO[0].BytesRecv
		netSpeed.UpLoad = newIO[0].BytesSent - oldIO[0].BytesSent
		if netSpeed.DownLoad == netSpeed.UpLoad {
			netSpeed = struct {
				DownLoad uint64
				UpLoad   uint64
			}{0, 0}
		}
		netMute.Unlock()
	}
}

// 返回当前系统时间
func getDate(format string) string {
	now := time.Now()
	return now.Format(format)
}

// 获取内存占用
func getMem(format string) string {
	memInfo, _ := mem.VirtualMemory()
	memUsed, unit := util.CountSize(memInfo.Total - memInfo.Available)
	return fmt.Sprintf(format+"%s", memUsed, unit)
}

// 获取硬盘占用
func getDisk(format string, mountPath string) string {
	parts, _ := disk.Partitions(true)
	for _, v := range parts {
		if v.Mountpoint == mountPath {
			diskInfo, _ := disk.Usage(v.Mountpoint)
			result, unit := util.CountSize(diskInfo.Free)
			return fmt.Sprintf(format+"%s", result, unit)
		}
	}
	return ""
}

// 获取cpu温度
func getCpuTemp(format string) string {
	byte, _ := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	temp, _ := strconv.ParseUint(strings.ReplaceAll(string(byte), "\n", ""), 10, 64)
	return fmt.Sprintf(format, temp/1000)
}

func getBright(format string) string {
	maxByte, _ := ioutil.ReadFile("/sys/class/backlight/intel_backlight/max_brightness")
	nowByte, _ := ioutil.ReadFile("/sys/class/backlight/intel_backlight/brightness")
	max, _ := strconv.ParseUint(strings.ReplaceAll(string(maxByte), "\n", ""), 10, 64)
	now, _ := strconv.ParseUint(strings.ReplaceAll(string(nowByte), "\n", ""), 10, 64)
	return fmt.Sprintf(format, float64(now)*100/float64(max))
}

func getNote(indexChar string) string {
	u, _ := user.Current()
	byte, _ := ioutil.ReadFile(fmt.Sprintf("%s/.note", u.HomeDir))
	flags := strings.Split(string(byte), "\n")
	validFlags := []string{}
	for i, v := range flags {
		if strings.Index(v, indexChar) > -1 {
			validFlags = append(validFlags, fmt.Sprintf("%d.%s", i, strings.TrimSpace(v[1:])))
		}
	}
	return strings.Join(validFlags, " | ")
}
