package main

func healthStatus(percentage float64) string {
	if percentage <= 70 {
		return "[OK]"
	} else if percentage <= 85 {
		return "[WARNING]"
	} else {
		return "[CRITICAL]"
	}
}

func latencyStatus(latency float64) string {
	if latency <= 300 {
		return "[OK]"
	} else if latency <= 1000 {
		return "[WARNING]"
	} else {
		return "[CRITICAL]"
	}
}

func tempStatus(temp float64) string {
	if temp <= 70 {
		return "[OK]"
	} else if temp <= 85 {
		return "[WARNING]"
	} else {
		return "[CRITICAL]"
	}
}

func countStatus(count int) string {
	if count == 0 {
		return "[OK]"
	} else if count >= 1 && count <= 5 {
		return "[WARNING]"

	} else {
		return "[CRITICAL]"
	}
}

func networkStatus(status bool) string {
	if status {
		return "[OK]"
	}
	return "[FAILED]"
}

func overAllHealth(statuses []string) string {
	overall := "[OK]"
	for _, status := range statuses {
		switch status {
		case "[CRITICAL]":
			return "[CRITICAL]"
		case "[ERROR]":
			overall = "[ERROR]"
		case "[WARNING]":
			if overall == "[OK]" {
				overall = "[WARNING]"
			}
		}
	}
	return overall
}
