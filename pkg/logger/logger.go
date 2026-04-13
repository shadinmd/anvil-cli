package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var Logger *log.Logger

func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	logDir := filepath.Join(home, ".config", "anvil")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile := filepath.Join(logDir, "anvil.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	Logger = log.New(f, "ANVIL: ", log.LstdFlags|log.Lshortfile)
	Logger.Println("Logger initialized")
	return nil
}

func Log(v ...interface{}) {
	if Logger != nil {
		Logger.Println(v...)
	}
}

func Logf(format string, v ...interface{}) {
	if Logger != nil {
		Logger.Printf(format, v...)
	}
}
