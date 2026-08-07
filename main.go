package main
import (
	"fmt"
	"log"
	"os"
	"strings"
)

func checkArch() bool {
	content, err := os.ReadFile("/etc/os-release")
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}
	archOrNo := strings.Contains(string(content), "ID=arch")
	
	return archOrNo
	
}

func main() {
	isArch := checkArch()
	if isArch {
		fmt.Println("Arch Linux: OK")
	} else {
		fmt.Println("Arch Linux: FAILED")
	}

}
