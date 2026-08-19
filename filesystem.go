package main

import (
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/sys/unix"
)

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
