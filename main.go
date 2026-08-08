package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
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

func checkInternet() bool {
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get("https://google.com")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return true
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
		fmt.Println("  Failed Serices: 0")
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

	fmt.Printf("  Disk /: %.1f%% used\n", perctangeUsed)

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
				fmt.Println("  Error: ", err)
				return
			}
			availableMemory = value
		}
	}
	memoryUsed := totalMemory - availableMemory
	perctangeUsedMem := (memoryUsed / totalMemory) * 100
	fmt.Printf("  Memory: %.1f%% used \n", perctangeUsedMem)

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
					fmt.Println("  Error: ", err)
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
		fmt.Println("CPU usage: FAILED")
		return
	}
	idleChange := idle2 - idle1
	cpuUsage := (1 - (idleChange / totalChange)) * 100

	fmt.Printf("  CPU usage: %.1f%% \n", cpuUsage)

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

	isInternet := checkInternet()
	if isInternet {
		fmt.Println("  Internet: OK")
	} else {
		fmt.Println("  Internet: Failed")
	}

	fmt.Println()

	fmt.Println("Health")

	checkCPU()
	checkDisk("/")
	checkMemory()
	checkFailedServices()
	fmt.Println()

	fmt.Println("Packages")
	checkUpdates()

}
