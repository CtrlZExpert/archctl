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

func runNetwork() {
	interfaceName := getActiveInterface()
	ipv4 := getIPv4Address(interfaceName)
	defaultGateway := getDefaultGateway()
	dnsServer := getDNSServer()
	gatewayReachable := checkGateway(defaultGateway)
	internetOK, latency := checkInternet()
	dnsOK := checkDNSResolution()

	fmt.Println("Network:")
	fmt.Println("  Interface:", interfaceName)
	fmt.Println("  IPv4:", ipv4)
	fmt.Println("  Gateway:", defaultGateway)
	fmt.Println("  DNS Server:", dnsServer)
	fmt.Println()

	gatewayStatus := networkStatus(gatewayReachable)
	internetStatus := networkStatus(internetOK)
	dnsStatus := networkStatus(dnsOK)
	latencyState := latencyStatus(latency)
	fmt.Println("Connectivity:")
	fmt.Println("  Gateway:", gatewayStatus)
	fmt.Println("  Internet:", internetStatus)
	fmt.Println("  DNS Resolution:", dnsStatus)
	fmt.Printf("  Latency: %.1f ms %s\n", latency, latencyState)

	var actionChoice string

	for {
		fmt.Println()
		fmt.Println("Network Action:")
		fmt.Println("1. Ping gateway")
		fmt.Println("2. Test internet connection")
		fmt.Println("3. Test DNS resolution")
		fmt.Println("4. Lookup Hostname")
		fmt.Println("5. Exit")
		fmt.Println()
		fmt.Print("Select an action: ")

		fmt.Scan(&actionChoice)
		fmt.Println()
		validChoice, err := strconv.Atoi(actionChoice)
		if err != nil {
			fmt.Println("  Invalid selection. Enter a num from 1-5")
			continue
		}

		switch validChoice {
		case 1:
			fmt.Println("  Gateway:", networkStatus(checkGateway(defaultGateway)))
		case 2:
			internetOK, latency = checkInternet()
			fmt.Println("  Internet:", networkStatus(internetOK))
			fmt.Printf("  Latency: %.1f ms %s\n", latency, latencyStatus(latency))
		case 3:
			fmt.Println("  DNS Resolution:", networkStatus(checkDNSResolution()))
		case 4:
			var hostname string
			fmt.Print("Enter hostname: ")
			fmt.Scan(&hostname)
			fmt.Println()

			cmd := exec.Command(
				"getent",
				"hosts",
				hostname,
			)
			output, err := cmd.Output()
			if err != nil {
				fmt.Println("  Hostname lookup failed")
				continue
			}
			str := string(output)
			fmt.Printf("  %s", str)

		case 5:
			return
		}

	}

}
