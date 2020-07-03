package log

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	Debug  = 1 << iota // 1
	Info               // 2
	Notice             // 4

	Nothing = 1024
)

var logText = map[int]string{
	Debug:  "DEBUG",
	Info:   "INFO",
	Notice: "NOTICE",

	Nothing: "NOTHING",
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

func ParseLogLevel(logLevel string) int {
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		return Debug
	case "INFO":
		return Info
	case "NOTICE":
		return Notice
	default:
		return Nothing
	}
}
