package logger

import (
	"io"
	"log"
	"os"
)

var (
	Info  *log.Logger
	Error *log.Logger
	Warn  *log.Logger
)

func Init(logFilePath string) {
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("cannot open log file: %v", err)
	}
	flags := log.Ldate | log.Ltime | log.Lshortfile
	Info = log.New(io.MultiWriter(os.Stdout, f),
		"INFO: ", flags)

	Warn = log.New(io.MultiWriter(os.Stdout, f),
		"WARN: ", flags)

	Error = log.New(io.MultiWriter(os.Stderr, f),
		"ERROR: ", flags)
}
