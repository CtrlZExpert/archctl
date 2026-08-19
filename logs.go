package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

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

	var actionChoice string

	for {
		fmt.Println()
		fmt.Println("1. Show all recent error logs")
		fmt.Println("2. Show kernel errors")
		fmt.Println("3. Show errors by service")
		fmt.Println("4. Exit")
		fmt.Println()
		fmt.Print("Select an action: ")
		

		fmt.Scan(&actionChoice)
		fmt.Println()
		validChoice, err := strconv.Atoi(actionChoice)
		if err != nil {
			fmt.Println("  Invalid selection. Enter a num from 1-4")
			continue
		}
				

		switch validChoice {
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

func checkJournalErrors() string {
	journalErrors := getJournalErrors()
	count := len(journalErrors)
	status := countStatus(count)
	fmt.Printf("  Journal Errors (1h): %d %s\n", count, status)
	return status

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
