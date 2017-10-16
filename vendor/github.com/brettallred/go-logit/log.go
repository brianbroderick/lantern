package logit

import (
	"errors"
	"log"
	"os"
	"strings"
	"sync/atomic"
)

const (
	// DEBUG doesn't filter any logs
	DEBUG int32 = 0
	// INFO filters only Debug logs
	INFO int32 = 1
	// WARN filters Info and Debug logs
	WARN int32 = 2
	// ERROR prints Error and Fatal logs
	ERROR int32 = 3
	// FATAL prints only Fatal logs
	FATAL int32 = 4
)

var (
	logLevel      int32
	errWrongLevel = errors.New("incorrect log level, options are DEBUG|INFO|WARN|ERROR|FATAL")
)

func init() {
	initialLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if initialLevel == "" {
		log.Println("Log level not provided in ENV. Setting to ERROR.")
		initialLevel = "ERROR"
	}
	err := setLogLevelByName(initialLevel)
	if err != nil {
		log.Println(err)
	}
}

// Debug prints the format string with the prefix of "DEBUG:"
func Debug(format string, v ...interface{}) {
	if LogLevel() <= 0 {
		log.Printf("DEBUG:"+format, v...)
	}
}

// Info prints the format string with the prefix of "INFO:"
func Info(format string, v ...interface{}) {
	if LogLevel() <= 1 {
		log.Printf("INFO:"+format, v...)
	}
}

// Warn prints the format string with the prefix of "WARN:"
func Warn(format string, v ...interface{}) {
	if LogLevel() <= 2 {
		log.Printf("WARN:"+format, v...)
	}
}

// Error prints the format string with the prefix of "ERROR:"
func Error(format string, v ...interface{}) {
	if LogLevel() <= 3 {
		log.Printf("ERROR:"+format, v...)
	}
}

// Fatal prints the format string with the prefix of "FATAL:"
func Fatal(format string, v ...interface{}) {
	if LogLevel() <= 4 {
		log.Printf("FATAL:"+format, v...)
	}
}

func setLogLevelByName(newLevel string) error {
	switch newLevel {
	case "DEBUG":
		SetLogLevel(0)
	case "INFO":
		SetLogLevel(1)
	case "WARN":
		SetLogLevel(2)
	case "ERROR":
		SetLogLevel(3)
	case "FATAL":
		SetLogLevel(4)
	default:
		return errWrongLevel
	}
	return nil
}

// SetLogLevel takes the logLevel int and changes the level in a thread-safe way
func SetLogLevel(newLevel int32) error {
	if newLevel < 0 || newLevel > 4 {
		return errWrongLevel
	}
	atomic.StoreInt32(&logLevel, newLevel)
	return nil
}

// LogLevelName reads the level's name in a thread-safe way
func LogLevelName() string {
	switch LogLevel() {
	case 0:
		return "DEBUG"
	case 1:
		return "INFO"
	case 2:
		return "WARN"
	case 3:
		return "ERROR"
	case 4:
		return "FATAL"
	default:
		return "Log level error"
	}
}

// LogLevel reads the level in a thread-safe way
func LogLevel() int32 {
	return atomic.LoadInt32(&logLevel)
}
