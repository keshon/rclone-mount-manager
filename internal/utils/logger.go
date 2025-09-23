package utils

import (
	"log"
	"os"
)

func FileLogger(path string) *log.Logger {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	return log.New(f, "", log.LstdFlags)
}
