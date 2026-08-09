package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

func checkArch() bool {
	content, err := os.ReadFile("/etc/os-release")
	if err != nil {
		log.Fatalf("  Failed to read file: %s", err)
	}
	archOrNo := strings.Contains(string(content), "ID=arch")

	return archOrNo

}

func checkInternet() (bool, int64) {
	start := time.Now()
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get("https://google.com")
	if err != nil {
		return false, 0
	}
	defer resp.Body.Close()
	elapsed := time.Since(start).Milliseconds()

	return true, elapsed
}

func checkUpdates() {
	path, err := exec.LookPath("checkupdates")
	if err != nil {
		fmt.Println("  Updates: FAILED (checkupdates not installed)")
		return
	}
	cmd := exec.Command(path)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("  Updates: FAILED (checkupdates not intstalled)")
		return
	}

	str := string(output)
	count := strings.Count(str, "\n")
	fmt.Println("  Updates: ", count)

}
func checkFailedServices() {
	cmd := exec.Command("systemctl", "--failed", "--no-legend")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("  Failed service: FAILED")
		return
	}
	str := string(output)
	str = strings.TrimSpace(str)
	if str == "" {
		fmt.Println("  Failed Services: 0")
		return
	}
	lines := strings.Split(str, "\n")
	count := len(lines)

	fmt.Println("  Failed Services: ", count)
}

func checkDisk(path string) {
	var stat unix.Statfs_t

	err := unix.Statfs(path, &stat)
	if err != nil {
		fmt.Println("  Disk: FAILED")
		return
	}
	totalSpace := stat.Blocks
	freeSpace := stat.Bfree
	availableSpace := stat.Bavail

	totalSpaceUsed := totalSpace - freeSpace
	usableSpace := totalSpace - freeSpace + availableSpace
	perctangeUsed := (float64(totalSpaceUsed) / float64(usableSpace)) * 100
	status := healthStatus(perctangeUsed)

	fmt.Printf("  Disk /: %.1f%% used %s\n", perctangeUsed, status)

}
func checkMemory() {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		log.Fatalf("  Failed to read file: %s", err)
	}

	var totalMemory float64
	var availableMemory float64

	str := string(content)
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			value, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				fmt.Println(" Error: ", err)
				return
			}
			totalMemory = value

		}
		if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			value, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				fmt.Println("  Error:", err)
				return
			}
			availableMemory = value
		}
	}
	memoryUsed := totalMemory - availableMemory
	perctangeUsedMem := (memoryUsed / totalMemory) * 100
	status := healthStatus(perctangeUsedMem)
	fmt.Printf("  Memory: %.1f%% used %s \n", perctangeUsedMem, status)

}

func readCPUStats() (totalTime float64, idleTime float64) {
	content, err := os.ReadFile("/proc/stat")
	if err != nil {
		log.Fatalf(" Failed to read file: %s", err)
	}
	var values []float64
	str := string(content)
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			for _, field := range fields[1:] {
				value, err := strconv.ParseFloat(field, 64)
				if err != nil {
					fmt.Println("  Error:", err)
					return
				}
				values = append(values, value)
			}
		}
	}
	for _, value := range values {
		totalTime += value
	}
	idle := values[3]
	iowait := values[4]
	idleTime = idle + iowait
	return totalTime, idleTime
}

func checkCPU() {

	total1, idle1 := readCPUStats()
	time.Sleep(1 * time.Second)
	total2, idle2 := readCPUStats()

	totalChange := total2 - total1
	if totalChange == 0 {
		fmt.Println("  CPU usage: FAILED")
		return
	}
	idleChange := idle2 - idle1
	cpuUsage := (1 - (idleChange / totalChange)) * 100
	status := healthStatus(cpuUsage)

	fmt.Printf("  CPU usage: %.1f%% %s\n", cpuUsage, status)

}

func healthStatus(percentage float64) string {
	if percentage <= 70 {
		return "[OK]"
	} else if percentage <= 85 {
		return "[WARNING]"
	} else {
		return "[CRITICAL]"
	}
}

func latencyStatus(latency int64) string {
	if latency <= 300 {
		return "[OK]"
	} else if latency <= 1000 {
		return "[WARNING]"
	} else {
		return "[CRITICAL]"
	}
}

func checkUptime() {
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		log.Fatalf("  Failed to read file: %s", err)
		return
	}
	str := string(content)
	fields := strings.Fields(str)
	totalSeconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		fmt.Println(" Error: ", err)
		return
	}
	days := int64(totalSeconds) / 86400
	remainingSeconds := int64(totalSeconds) % 86400
	hours := remainingSeconds / 3600
	remainingSeconds = remainingSeconds % 3600
	minutes := remainingSeconds / 60

	fmt.Printf("  Uptime: %d days %d hours %d minutes \n", days, hours, minutes)
}

func checkAverageLoad() {
	content, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		log.Fatalf("  Failed to read file: %s", err)
		return
	}
	str := string(content)
	fields := strings.Fields(str)
	load1, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		fmt.Println("  Error:", err)
		return
	}

	load2, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		fmt.Println("  Error:", err)
		return
	}
	load3, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		fmt.Println("  Error:", err)
		return
	}
	cpuCount := runtime.NumCPU()

	loadPercentage := (load1 / float64(cpuCount)) * 100
	status := healthStatus(loadPercentage)

	fmt.Printf("  Load Average: %.1f %.1f %.1f %s\n", load1, load2, load3, status)

}

func checkSwap() {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		log.Fatalf(" Failed to read file: %s ", err)
		return
	}
	var swapTotal float64
	var swapFree float64
	str := string(content)
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if strings.HasPrefix(line, "SwapTotal:") {
			value, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				fmt.Println(" Error: ", err)
				return
			}
			swapTotal = value
		}
		if strings.HasPrefix(line, "SwapFree:") {
			value, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				fmt.Println("  Error:", err)
				return
			}
			swapFree = value
		}
	}
	if swapTotal == 0 {
		fmt.Println("  Swap: Not configured")
		return
	}

	usedSwap := swapTotal - swapFree
	swapPerctangeUsed := (usedSwap / swapTotal) * 100
	status := healthStatus(swapPerctangeUsed)

	fmt.Printf("  Swap %.1f%% used %s\n", swapPerctangeUsed, status)

}
func getRunningKernel() (string, error) {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	str := string(output)
	return str, nil
}

func checkRebootRequired() bool {
	unameStr, err := getRunningKernel()
	if err != nil {
		fmt.Println("  Error:", err)
		return false
	}

	unameKernel := strings.TrimSpace(unameStr)

	pacmanCmd := exec.Command("pacman", "-Q", "linux")
	pacmanOutput, err := pacmanCmd.Output()
	if err != nil {
		fmt.Println(" Error:", err)
		return false
	}
	pacmanStr := string(pacmanOutput)
	pacmanFields := strings.Fields(pacmanStr)
	pacmanField := pacmanFields[1]
	pacmanKernel := strings.Replace(pacmanField, ".arch", "-arch", 1)
	if unameKernel != pacmanKernel {
		return true
	}

	return false
}

func main() {

	fmt.Println("archctl - Arch Linux System Doctor")
	fmt.Println("____________________________________")
	fmt.Println()
	fmt.Println()

	fmt.Println("System")
	isArch := checkArch()
	if isArch {
		fmt.Println("  Arch Linux: OK")
	} else {
		fmt.Println("  Arch Linux: FAILED")
	}
	currentKernel, err := getRunningKernel()
	if err != nil {
		fmt.Println(" Error:", err)
		return
	}
	fmt.Println("  Kernel:", currentKernel)

	isInternet, elapsed := checkInternet()
	if isInternet {
		status := latencyStatus(elapsed)
		fmt.Println("  Internet: OK")
		fmt.Printf("  Latency: %d ms %s\n", elapsed, status)
	} else {
		fmt.Println("  Internet: Failed")
	}

	fmt.Println()

	fmt.Println("Health")

	checkCPU()
	checkAverageLoad()
	checkDisk("/")
	checkMemory()
	checkSwap()
	checkFailedServices()
	checkUptime()
	fmt.Println()

	fmt.Println("Packages")
	checkUpdates()
	rebootRequired := checkRebootRequired()
	if rebootRequired {
		fmt.Println("  Reboot Required: Yes")
	} else {
		fmt.Println("  Reboot Required: No")
	}

}
