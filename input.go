package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

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
