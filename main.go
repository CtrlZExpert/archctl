package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

func checkUpdates() int {
	path, err := exec.LookPath("checkupdates")
	if err != nil {
		fmt.Println("  Updates: FAILED (checkupdates is not available)")
		return 0
	}
	cmd := exec.Command(path)
	output, err := cmd.Output()
	str := string(output)
	updateOutput := strings.TrimSpace(str)

	if updateOutput == "" {
		fmt.Println("  Updates: 0")
		return 0
	}
	if err != nil {
		fmt.Println("  Updates: FAILED (checkupdates failed to run)")
		return 0
	}
	updates := strings.Split(updateOutput, "\n")

	count := len(updates)
	fmt.Println("  Updates: ", count)
	return count

}

func checkFailedServices() (string, []string) {
	cmd := exec.Command("systemctl", "--failed", "--no-legend", "--plain", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("  Failed service: FAILED")
		return "[ERROR]", []string{}
	}
	str := string(output)
	str = strings.TrimSpace(str)
	if str == "" {
		return "[OK]", []string{}
	}
	lines := strings.Split(str, "\n")
	var serviceNames []string

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			serviceNames = append(serviceNames, fields[0])
		}
	}
	count := len(serviceNames)
	status := countStatus(count)
	return status, serviceNames
}

func checkDisk(path string) string {
	var stat unix.Statfs_t

	err := unix.Statfs(path, &stat)
	if err != nil {
		fmt.Println("  Disk: FAILED")
		return "[ERROR]"
	}
	totalSpace := stat.Blocks
	freeSpace := stat.Bfree
	availableSpace := stat.Bavail

	totalSpaceUsed := totalSpace - freeSpace
	usableSpace := totalSpace - freeSpace + availableSpace
	perctangeUsed := (float64(totalSpaceUsed) / float64(usableSpace)) * 100
	status := healthStatus(perctangeUsed)

	fmt.Printf("  Disk %s: %.1f%% used %s\n", path, perctangeUsed, status)
	return status

}

func getFileSystem() {
	cmd := exec.Command("findmnt", "-rn", "-o", "TARGET,SOURCE,FSTYPE")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("  Error:", err)
		return
	}
	var rootTarget string
	var homeTarget string
	var bootTarget string
	var rootSource string
	var homeSource string
	var bootSource string
	var rootFSType string
	var homeFSType string
	var bootFSType string

	outputStr := string(output)

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "/" {
			rootTarget = fields[0]
			rootSource = fields[1]
			rootFSType = fields[2]
		}
		if fields[0] == "/home" {
			homeTarget = fields[0]
			homeSource = fields[1]
			homeFSType = fields[2]

		}
		if fields[0] == "/boot" {
			bootTarget = fields[0]
			bootSource = fields[1]
			bootFSType = fields[2]

		}

		if rootFSType == "btrfs" {
			rootSource = strings.Replace(rootSource, "[/@]", "", 1)
			homeSource = strings.Replace(homeSource, "[/@home]", "", 1)
		}
	}

	fmt.Printf("  %s %s %s\n", rootTarget, rootSource, rootFSType)
	checkDisk("/")

	if rootSource != homeSource {
		fmt.Printf("  %s %s %s\n", homeTarget, homeSource, homeFSType)
		checkDisk("/home")
	}

	fmt.Printf("  %s %s %s\n", bootTarget, bootSource, bootFSType)
	checkDisk("/boot")

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

	fmt.Printf("  Swap %.1f%% used %s\n", swapPerctangeUsed, status)
	return status

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

func tempStatus(temp float64) string {
	if temp <= 70 {
		return "[OK]"
	} else if temp <= 85 {
		return "[WARNING]"
	} else {
		return "[CRITICAL]"
	}
}

func checkJournalErrors() string {
	journalErrors := getJournalErrors()
	count := len(journalErrors)
	status := countStatus(count)
	fmt.Printf("  Journal Errors (1h): %d %s\n", count, status)
	return status

}

func getJournalErrors() []string {
	cmd := exec.Command(
		"journalctl",
		"-p",
		"err",
		"--since",
		"1 hour ago",
		"--no-pager",
	)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("  Error:", err)
		return []string{}
	}
	journalStr := strings.TrimSpace(string(output))
	if journalStr == "" {
		return []string{}
	}
	journalSet := strings.Split(journalStr, "\n")

	return journalSet

}
func getKernelErrors() []string {
	cmd := exec.Command(
		"journalctl",
		"-k",
		"-p",
		"err",
		"--since",
		"1 hour ago",
		"--no-pager",
		"--quiet",
	)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("  Error:", err)
		return []string{}
	}
	kernelStr := strings.TrimSpace(string(output))
	if kernelStr == "" {
		return []string{}
	}
	kernelSet := strings.Split(kernelStr, "\n")

	return kernelSet

}

func countStatus(count int) string {
	if count == 0 {
		return "[OK]"
	} else if count >= 1 && count <= 5 {
		return "[WARNING]"

	} else {
		return "[CRITICAL]"
	}
}

func overAllHealth(statuses []string) string {
	overall := "[OK]"
	for _, status := range statuses {
		switch status {
		case "[CRITICAL]":
			return "[CRITICAL]"
		case "[ERROR]":
			overall = "[ERROR]"
		case "[WARNING]":
			if overall == "[OK]" {
				overall = "[WARNING]"
			}
		}
	}
	return overall
}
func runDoctor() {

	runSystem()
	fmt.Println()
	fmt.Println()
	fmt.Println("Filesytem")
	getFileSystem()
	fmt.Println()
	printHealth()
	fmt.Println()
	runPackages()
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

func printHealth() {
	statuses := runHealth()
	overall := overAllHealth(statuses)
	fmt.Println("Overall Health:", overall)

}
func runSystem() {
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

}

func runPackages() {
	fmt.Println("Packages")
	checkUpdates()
	rebootRequired := checkRebootRequired()
	if rebootRequired {
		fmt.Println("  Reboot Required: Yes")
	} else {
		fmt.Println("  Reboot Required: No")
	}

}
func runUpdates() {
	fmt.Println("Sytem Update")
	numUpdates := checkUpdates()
	if numUpdates == 0 {
		fmt.Println("  System is up to date")
		return
	}
	prompt := "  Would you like to update your system? (yes/no): "
	confirmUpdate := confirmAction(prompt)
	if confirmUpdate {
		cmd := exec.Command("sudo", "pacman", "-Syu")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println("  Error:", err)
			return
		}
		fmt.Println("  System update completed successfully")
		runPackages()
		return
	}
	fmt.Println("  Update cancelled")
	return

}

func runClean() {
	cmd := exec.Command("pacman", "-Qdtq")
	orphanOutput, err := cmd.Output()
	orphanStr := string(orphanOutput)
	orphanPackages := strings.TrimSpace(orphanStr)

	if orphanPackages == "" {
		fmt.Println("  No orphan packages found")
		return
	}
	if err != nil {
		fmt.Println("  Error:", err)
		return
	}

	fmt.Println(orphanPackages)
	fmt.Println()
	fmt.Println("Warning: orphan packages may still be manually useful")
	fmt.Println("Review the list before confirming removal")
	fmt.Println()

	var packagesToRemove []string
	orphanNames := strings.Split(orphanPackages, "\n")
	orphanSet := make(map[string]bool)
	var packagesToKeep []string
	for _, name := range orphanNames {
		orphanSet[name] = true
	}
	for {
		packagesToKeep = selectPackagesToKeep()
		valid := true

		for _, item := range packagesToKeep {
			if !orphanSet[item] {
				fmt.Printf("Error: %q is not an orphan package\n", item)
				valid = false
			}
		}
		if valid {
			break
		}

	}
	for _, pkg := range orphanNames {
		keep := false
		for _, keepPackage := range packagesToKeep {
			if pkg == keepPackage {
				keep = true
				break
			}
		}
		if !keep {
			packagesToRemove = append(packagesToRemove, pkg)
		}

	}

	if len(packagesToRemove) == 0 {
		fmt.Println("  Cleanup cancelled")
		return
	}
	fmt.Println("Packages to remove:")
	for _, removePkg := range packagesToRemove {
		fmt.Println(removePkg)
	}
	fmt.Println()
	args := append([]string{"pacman", "-Rns"}, packagesToRemove...)

	prompt := "Remove these orphan packages? (yes/no):  "
	confirmClean := confirmAction(prompt)
	if confirmClean {
		cmd := exec.Command("sudo", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println("  Error:", err)
			return
		}
		fmt.Println("  Packages have successfully been removed")
		return

	}
	fmt.Println("  Cleanup cancelled")
}

func runService() {
servicesLoop:
	for {
		_, failedServices := checkFailedServices()
		numFailedServices := len(failedServices)
		if numFailedServices == 0 {
			fmt.Println("No failed service found")
			return
		}
		fmt.Println("Failed Services:")
		for i, service := range failedServices {
			fmt.Printf("  %d.  %s\n", i+1, service)
		}
		var usrOptions int
		for {
			fmt.Print("Select a service by number: ")
			fmt.Println()

			fmt.Scan(&usrOptions)
			if usrOptions < 1 || usrOptions > numFailedServices {
				fmt.Println("Invalid selection. Please enter a number")
				fmt.Println()
				continue
			}
			break
		}

		var actionChoice int

		for {
			selectedService := failedServices[usrOptions-1]
			fmt.Printf("Selected: %s\n", selectedService)
			fmt.Println()
			fmt.Println("  1. View status")
			fmt.Println("  2. View recent logs")
			fmt.Println("  3. Restart service")
			fmt.Println("  4. Exit")
			fmt.Println()

			fmt.Print("Select an action: ")
			fmt.Scan(&actionChoice)

			switch actionChoice {
			case 1:
				cmd := exec.Command(
					"systemctl", "status", selectedService)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					var exitErr *exec.ExitError
					if !errors.As(err, &exitErr) {
						fmt.Println("Error running systemctl:", err)
					}
				}
			case 2:
				cmd := exec.Command(
					"journalctl",
					"-u", selectedService,
					"--since",
					"1 hour ago",
				)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					fmt.Println("Error running journalctl:", err)
				}
			case 3:
				prompt := fmt.Sprintf("Restart %s? (yes/no): ", selectedService)
				confirmRestart := confirmAction(prompt)
				if confirmRestart {
					cmd := exec.Command("sudo", "systemctl", "restart", selectedService)
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err := cmd.Run()
					if err != nil {
						fmt.Println("Error:", err)
						fmt.Println()
						continue servicesLoop
					}
					fmt.Println("Successfully restarted", selectedService)

				}
				if !confirmRestart {
					continue
				}

			case 4:
				return
			default:
				fmt.Println("Invalid section")
				fmt.Println()
			}

		}

	}
}

func runLog() {
	logErrors := getJournalErrors()
	numErrors := len(logErrors)
	if numErrors == 0 {
		fmt.Println("No journal errors found in the last hour")
		return
	}

	fmt.Println("Journal Errors - Last Hour:")
	for i, logEntry := range logErrors {
		fmt.Printf("  %d. %s\n", i+1, logEntry)
	}

	var actionChoice int

	for {
		fmt.Println()
		fmt.Println("1. Show all recent error logs")
		fmt.Println("2. Show kernel errors")
		fmt.Println("3. Show errors by service")
		fmt.Println("4. Exit")
		fmt.Println()
		fmt.Print("Select an action: ")
		fmt.Println()

		fmt.Scan(&actionChoice)

		switch actionChoice {
		case 1:
			fmt.Println("Recent Error Logs:")
			for i, logEntry := range logErrors {
				fmt.Printf("  %d. %s\n", i+1, logEntry)
			}
		case 2:
			kernelErrs := getKernelErrors()
			numKernelErrs := len(kernelErrs)
			fmt.Println("Kernel Errors - Last Hour:")
			if numKernelErrs == 0 {
				fmt.Println("  No kernel errors in the last hour ")
				continue
			}
			for i, kernelEntry := range kernelErrs {
				fmt.Printf("  %d. %s\n", i+1, kernelEntry)
			}
		case 3:
			fmt.Print("Enter service name: ")
			var serviceName string
			fmt.Scan(&serviceName)
			fmt.Println("Error by Service:")
			cmd := exec.Command(
				"journalctl",
				"-u",
				serviceName,
				"-p",
				"err",
				"--since",
				"1 hour ago",
				"--no-pager",
				"--quiet",
			)
			output, err := cmd.Output()
			if err != nil {
				fmt.Println("  Error:", err)
				continue
			}
			str := string(output)
			service := strings.TrimSpace(str)
			if service == "" {
				fmt.Printf("  No error logs found for %s in the last hour\n", serviceName)
				continue
			}
			fmt.Printf("  %s\n", service)
		case 4:
			return
		default:
			fmt.Println("Invalid selection")
		}
	}
}

func confirmAction(prompt string) bool {
	fmt.Print(prompt)
	for {
		var answer string
		fmt.Scan(&answer)
		answer = strings.ToLower(answer)
		if answer == "yes" || answer == "y" {
			return true
		}

		if answer == "no" || answer == "n" {
			return false
		}
		if answer != "yes" && answer != "y" && answer != "no" && answer != "n" {
			fmt.Println()
			fmt.Printf("Error: unexpected answer %q found\n", answer)
			fmt.Println()
			fmt.Print("  (y)es / (n)o: ")
			fmt.Println()
		}

	}

}

func selectPackagesToKeep() []string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter orphan packages you to keep (space-separated):")
	fmt.Println("Press Enter to remove all:")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("  Error reading input:", err)
		return []string{}
	}

	input = strings.TrimSpace(input)
	packages := strings.Fields(input)
	return packages
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  archctl <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  doctor    Run full system health chehck")
	fmt.Println("  health    Run health check")
	fmt.Println("  system    Show system information")
	fmt.Println("  packages  Show package/update status")
	fmt.Println("  update    Update system")
	fmt.Println("  clean     Find and remove orphan packages")
	fmt.Println("  services  Show failed systemd services")
	fmt.Println("  log       View and investigate recent journal errors")
	fmt.Println("Options:")
	fmt.Println("  --help    Show this help message")
	fmt.Println(" --version Show archctl version")

}

func printVersion() {
	fmt.Println("archctl v.0.1.0")
}

func main() {
	fmt.Println("archctl - Arch Linux System Doctor")
	fmt.Println("____________________________________")
	fmt.Println()
	fmt.Println()

	if len(os.Args) < 2 {

		fmt.Println("Error: missing command")
		fmt.Println()
		fmt.Println("Usage: archctl <command>")
		fmt.Println()
		fmt.Println("For more information, try --help")
		return
	}

	command := os.Args[1]
	switch command {
	case "doctor":
		runDoctor()
	case "health":
		printHealth()
	case "system":
		runSystem()
	case "packages":
		runPackages()
	case "update":
		runUpdates()
	case "clean":
		runClean()
	case "services":
		runService()
	case "log":
		runLog()
	case "--help":
		printHelp()
	case "--version":
		printVersion()

	default:
		fmt.Printf("Error: command %q not found\n", command)
		fmt.Println()
		fmt.Println("Run 'archctl help' for available commands")

	}

}
