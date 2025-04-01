package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
)

var ErrEmptyLogFileName = errors.New("empty log file name")

func NewLogger(fileName string) (*log.Logger, error) {
	if fileName == "" {
		return nil, ErrEmptyLogFileName
	}

	logFile, err := os.Create(fileName)
	if err != nil {
		return nil, fmt.Errorf("creating log file: %w", err)
	}

	return log.New(logFile, "", log.LstdFlags), nil
}
