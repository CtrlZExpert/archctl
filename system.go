package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
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
		fmt.Printf("  Latency: %.1f ms %s\n", elapsed, status)
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

func getRunningKernel() (string, error) {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	str := string(output)
	return str, nil
}
