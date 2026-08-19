package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func getActiveInterface() string {
	cmd := exec.Command(
		"ip",
		"route",
		"show",
		"default",
	)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error: ", err)
		return ""
	}

	str := string(output)
	fields := strings.Fields(str)
	for i, field := range fields {
		if field == "dev" && i+1 < len(fields) {
			return fields[i+1]

		}
	}
	return ""
}

func getIPv4Address(interfaceName string) string {
	cmd := exec.Command(
		"ip",
		"-4",
		"-o",
		"addr",
		"show",
		"dev",
		interfaceName,
		"scope",
		"global",
	)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	str := string(output)
	fields := strings.Fields(str)
	for i, field := range fields {
		if field == "inet" && i+1 < len(fields) {
			return fields[i+1]
		}
	}
	return ""
}

func getDefaultGateway() string {
	cmd := exec.Command(
		"ip",
		"route",
		"show",
		"default",
	)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	str := string(output)
	fields := strings.Fields(str)
	for i, field := range fields {
		if field == "via" && i+1 < len(fields) {
			return fields[i+1]
		}
	}
	return ""

}

func getDNSServer() string {
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	str := string(data)
	fields := strings.Fields(str)
	for i, field := range fields {
		if field == "nameserver" && i+1 < len(fields) {

			return fields[i+1]
		}
	}
	return ""
}

func checkGateway(gateway string) bool {
	cmd := exec.Command(
		"ping",
		"-c",
		"1",
		"-W",
		"2",
		gateway,
	)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func checkDNSResolution() bool {
	cmd := exec.Command(
		"getent",
		"hosts",
		"google.com",
	)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func checkInternet() (bool, float64) {
	cmd := exec.Command(
		"ping",
		"-c",
		"1",
		"-W",
		"2",
		"8.8.8.8",
	)
	output, err := cmd.Output()
	if err != nil {
		return false, 0.0
	}

	str := string(output)
	fields := strings.Fields(str)
	for _, field := range fields {
		if strings.HasPrefix(field, "time=") {
			timeField := strings.Replace(field, "time=", "", 1)

			elapsed, err := strconv.ParseFloat(timeField, 64)
			if err != nil {
				return false, 0.0
			}

			return true, elapsed
		}
	}
	return false, 0.0
}
