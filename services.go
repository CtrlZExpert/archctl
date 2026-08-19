package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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
