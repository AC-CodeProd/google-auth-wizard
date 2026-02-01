package logger

import (
	"fmt"
	"log"
	"os"
)

type LogLevel int

const (
	LEVEL_SILENT LogLevel = iota
	LEVEL_ERROR
	LEVEL_INFO
	LEVEL_DEBUG
	LEVEL_VERBOSE
)

type Logger struct {
	level  LogLevel
	prefix string
}

var globalLogger *Logger

func init() {
	initializeLogger()
}

func initializeLogger() {
	globalLogger = &Logger{
		level:  LEVEL_INFO,
		prefix: "[Google Auth Wizard] ",
	}

	if debug := os.Getenv("GOOGLE_AUTH_WIZARD_DEBUG"); debug == "true" || debug == "1" {
		globalLogger.level = LEVEL_DEBUG
	}

	if verbose := os.Getenv("GOOGLE_AUTH_WIZARD_VERBOSE"); verbose == "true" || verbose == "1" {
		globalLogger.level = LEVEL_VERBOSE
	}

	if silent := os.Getenv("GOOGLE_AUTH_WIZARD_SILENT"); silent == "true" || silent == "1" {
		globalLogger.level = LEVEL_SILENT
	}
}

func SetLevel(level LogLevel) {
	globalLogger.level = level
}

func GetLevel() LogLevel {
	return globalLogger.level
}

func Debug(format string, args ...interface{}) {
	if globalLogger.level >= LEVEL_DEBUG {
		log.Printf(globalLogger.prefix+"[DEBUG] "+format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if globalLogger.level >= LEVEL_INFO {
		log.Printf(globalLogger.prefix+"[INFO] "+format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if globalLogger.level >= LEVEL_ERROR {
		log.Printf(globalLogger.prefix+"[ERROR] "+format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	log.Fatalf(globalLogger.prefix+"[FATAL] "+format, args...)
}

func Print(format string, args ...interface{}) {
	if globalLogger.level > LEVEL_SILENT {
		fmt.Printf(format, args...)
	}
}

func Println(args ...interface{}) {
	if globalLogger.level > LEVEL_SILENT {
		fmt.Println(args...)
	}
}

func IsDebug() bool {
	return globalLogger.level >= LEVEL_DEBUG
}

func IsVerbose() bool {
	return globalLogger.level >= LEVEL_VERBOSE
}
