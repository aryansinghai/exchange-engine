package logger

import (
	"log"
	"time"
)

func Info(message string) {
	// add timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Println("[INFO] " + timestamp + " " + message)
}

func Error(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Println("[ERROR] " + timestamp + " " + message)
}

func Debug(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Println("[DEBUG] " + timestamp + " " + message)
}

func Warn(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Println("[WARN] " + timestamp + " " + message)
}
