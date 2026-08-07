package main
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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

}
