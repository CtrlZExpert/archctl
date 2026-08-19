package main

import (
	"fmt"
	"os"
)

func runDoctor() {

	runSystem()
	fmt.Println()
	fmt.Println()
	fmt.Println("Filesystem")
	getFileSystem()
	fmt.Println()
	printHealth()
	fmt.Println()
	runPackages()
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  archctl <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  doctor    Run full system health check")
	fmt.Println("  health    Run health check")
	fmt.Println("  system    Show system information")
	fmt.Println("  packages  Show package/update status")
	fmt.Println("  update    Update system")
	fmt.Println("  clean     Find and remove orphan packages")
	fmt.Println("  services  Show failed systemd services")
	fmt.Println("  log       View and investigate recent journal errors")
	fmt.Println("Options:")
	fmt.Println("  --help    Show this help message")
	fmt.Println("  --version Show archctl version")

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
		fmt.Println("Run 'archctl --help' for available commands")

	}

}
