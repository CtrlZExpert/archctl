package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func printHealth() {
	statuses := runHealth()
	overall := overAllHealth(statuses)
	fmt.Println("Overall Health:", overall)
}

func runHealth() []string {

	fmt.Println("Health")
	cpuStatus := checkCPU()
	tempStatus := checkTemp()
	loadStatus := checkAverageLoad()
	diskStatus := checkDisk("/")
	memoryStatus := checkMemory()
	swapStatus := checkSwap()
	failedServiceStatus, failedServices := checkFailedServices()
	numServices := len(failedServices)
	fmt.Printf("  Failed Services: %d %s\n", numServices, failedServiceStatus)
	jounrnalStatus := checkJournalErrors()
	checkUptime()
	fmt.Println()
	statuses := []string{
		cpuStatus,
		tempStatus,
		memoryStatus,
		swapStatus,
		loadStatus,
		diskStatus,
		failedServiceStatus,
		jounrnalStatus,
	}
	return statuses
}

func checkCPU() string {

	total1, idle1 := readCPUStats()
	time.Sleep(1 * time.Second)
	total2, idle2 := readCPUStats()

	totalChange := total2 - total1
	if totalChange == 0 {
		fmt.Println("  CPU usage: FAILED")
		return "[ERROR]"
	}
	idleChange := idle2 - idle1
	cpuUsage := (1 - (idleChange / totalChange)) * 100
	status := healthStatus(cpuUsage)

	fmt.Printf("  CPU usage: %.1f%% %s\n", cpuUsage, status)
	return status

}

func readCPUStats() (totalTime float64, idleTime float64) {
	content, err := os.ReadFile("/proc/stat")
	if err != nil {
		fmt.Println("  Failed to read file: ", err)
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

func checkTemp() string {
	paths, err := filepath.Glob("/sys/class/hwmon/hwmon*")
	if err != nil {
		fmt.Println("  Error:", err)
		return "[ERROR]"
	}

	for _, path := range paths {
		namePath := filepath.Join(path, "name")

		content, err := os.ReadFile(namePath)
		if err != nil {
			fmt.Println("  Error:", err)
			return "[ERROR]"
		}
		name := strings.TrimSpace(string(content))
		if name == "coretemp" {
			tempPath := filepath.Join(path, "temp1_input")
			tempContent, err := os.ReadFile(tempPath)
			if err != nil {
				fmt.Println(" Error:", err)
				return "[ERROR]"
			}
			tempStr := strings.TrimSpace(string(tempContent))
			temp, err := strconv.ParseFloat(tempStr, 64)
			if err != nil {
				fmt.Println("  Error:", err)
				return "[ERROR]"
			}
			temp = temp / 1000
			status := tempStatus(temp)
			fmt.Printf("  CPU Temp: %v°C  %s\n", temp, status)
			return status
		}
	}
	fmt.Println("  CPU Temp: sensor not found")
	return "[ERROR]"
}

func checkAverageLoad() string {
	content, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		fmt.Println("  Error:", err)
		return "[ERROR]"
	}
	str := string(content)
	fields := strings.Fields(str)
	load1, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		fmt.Println("  Error:", err)
		return "[ERROR]"
	}

	load2, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		fmt.Println("  Error:", err)
		return "[ERROR]"
	}
	load3, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		fmt.Println("  Error:", err)
		return "[ERROR]"
	}
	cpuCount := runtime.NumCPU()

	loadPercentage := (load1 / float64(cpuCount)) * 100
	status := healthStatus(loadPercentage)

	fmt.Printf("  Load Average: %.1f %.1f %.1f %s\n", load1, load2, load3, status)
	return status
}

func checkMemory() string {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		fmt.Println("  Error:", err)
		return "[ERROR]"
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
				return "[ERROR]"
			}
			totalMemory = value

		}
		if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			value, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				fmt.Println("  Error:", err)
				return "[ERROR]"
			}
			availableMemory = value
		}
	}
	if totalMemory == 0 {
		fmt.Println("  Error: Total Memory not found")
		return "[ERROR]"
	}
	memoryUsed := totalMemory - availableMemory
	perctangeUsedMem := (memoryUsed / totalMemory) * 100
	status := healthStatus(perctangeUsedMem)
	fmt.Printf("  Memory: %.1f%% used %s \n", perctangeUsedMem, status)
	return status

}

func checkSwap() string {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		fmt.Println("  Error:", err)
		return "[ERROR]"
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
				return "[ERROR]"
			}
			swapTotal = value
		}
		if strings.HasPrefix(line, "SwapFree:") {
			value, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				fmt.Println("  Error:", err)
				return "[ERROR]"
			}
			swapFree = value
		}
	}
	if swapTotal == 0 {
		fmt.Println("  Swap: Not configured")
		return "[ERROR]"
	}

	usedSwap := swapTotal - swapFree
	swapPerctangeUsed := (usedSwap / swapTotal) * 100
	status := healthStatus(swapPerctangeUsed)

	fmt.Printf("  Swap: %.1f%% used %s\n", swapPerctangeUsed, status)
	return status

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
