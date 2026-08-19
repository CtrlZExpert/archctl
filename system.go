package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

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

func getRunningKernel() (string, error) {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	str := string(output)
	return str, nil
}
