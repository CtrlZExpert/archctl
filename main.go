package main
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

func checkArch() bool {
	content, err := os.ReadFile("/etc/os-release")
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}
	archOrNo := strings.Contains(string(content), "ID=arch")
	
	return archOrNo
	
}


func checkInternet() bool {
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get("https://google.com")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return true
}

func checkUpdates() {
	path, err := exec.LookPath("checkupdates")
	if err != nil {
		fmt.Println("Updates: FAILED (checkupdates not installed)")
		return
		}
	cmd := exec.Command(path)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Updates: FAILED (checkupdtes not intstalled)")
		return
	}
	
	str := string(output)
	count := strings.Count(str, "\n")
	fmt.Println("Updates: ", count)
	

} 

func checkDisk(path string) {
	var stat unix.Statfs_t
	
	err := unix.Statfs(path, &stat)
	if err != nil {
		fmt.Println("Disk: FAILED")
		return
	}
	totalSpace := stat.Blocks
	freeSpace := stat.Bfree
	availableSpace := stat.Bavail

	totalSpaceUsed := totalSpace - freeSpace
	usableSpace:= totalSpace - freeSpace + availableSpace
	perctangeUsed:= (float64(totalSpaceUsed)/float64(usableSpace)) * 100

	fmt.Printf("Disk /: %.1f%% used\n",perctangeUsed)

}

func main() {
	isArch := checkArch()
	if isArch {
		fmt.Println("Arch Linux: OK")
	} else {
		fmt.Println("Arch Linux: FAILED")
	}

	isInternet := checkInternet()
	if isInternet {
		fmt.Println("Internet: OK")
	} else {
		fmt.Println("Internet: Failed")
	}

	checkUpdates()
	checkDisk("/")

}
