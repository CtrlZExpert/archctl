package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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
	fmt.Println("System Update")
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
