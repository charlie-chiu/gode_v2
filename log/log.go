package log

import (
	"fmt"
	"log"
	"os"
)

const (
	Debug = 1 << iota
	Info
	Notice

	Nothing
)

var logText = map[int]string{
	Debug:  "DEBUG",
	Info:   "INFO",
	Notice: "NOTICE",
}

var logger = log.New(os.Stderr, "", log.LstdFlags)

var level = Nothing

func init() {
	logger.SetFlags(log.Ltime)
}

func SetLevel(logLevel int) {
	level = logLevel
}

func Print(logLevel int, v ...interface{}) {
	if logLevel >= level {
		logger.SetPrefix(fmt.Sprintf("[%s]", logText[logLevel]))
		_ = logger.Output(2, fmt.Sprint(v...))
	}
}

func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}
