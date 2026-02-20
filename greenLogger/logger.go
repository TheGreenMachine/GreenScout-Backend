package greenlogger

// Logging wrapper to be used for error handling and general logging.

import (
	"GreenScoutBackend/constants"
	filemanager "GreenScoutBackend/fileManager"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// The reference to the current log file
var logFile *os.File

// If the log file is currently able to be written to. Prevents infinite recursions.
var logFileAlive bool

// Creates the log file and stores it to greenLogger/logger.go.logFile, setting logFileAlive to true.
// Panics if it is unable to create the file.
func InitLogFile() {
	filemanager.MkDirWithPermissions(constants.DefaultLogDirectory)
	logFilePath := filepath.Join(constants.DefaultLogDirectory, "GSLog_"+time.Now().Format("2006-01-02_h15m04s05")) //time format YEAR-MM-DD
	file, err := filemanager.OpenWithPermissions(logFilePath)
	if err != nil {
		panic("ERR: Could not create log file! " + err.Error())
	}

	logFile = file
	logFileAlive = true
}

// Logs an error and its message according to a format specifier to the console, log file, and slack.
// Params: The error, the message identifying that error as a format specifier, any args that fit into that format.
func LogErrorf(err error, message string, args ...any) {
	formatted := fmt.Sprintf(message, args...)
	LogError(err, formatted)
}

// Logs an error and its message to the console, log file, and slack.
// Params: The error, the message identifying that error
func LogError(err error, message string) {
	fmt.Println("ERR: " + message + ": " + err.Error())
	if constants.CachedConfigs.SlackConfigs.UsingSlack && slackAlive {
		NotifyError(err, message)
	}
	ElogError(err, message)
}

// Logs a message to the console and log file.
func LogMessage(message string) {
	fmt.Println(message)
	ELogMessage(message)
}

// Logs a message according to a format specifier to the console and log file.
// Params: The message as a format specifier, any args that fit into that format.
func LogMessagef(message string, args ...any) {
	formatted := fmt.Sprintf(message, args...)
	fmt.Println(formatted)
	ELogMessage(formatted)
}

// Exclusively logs a message to the log file
func ELogMessage(message string) {
	if logFileAlive {
		logFile.Write([]byte(time.Now().String() + ": " + message + "\n"))
	}
}

// Exclusively logs a message to the log file according to a format specifier
// Params: The message as a format specifier, any args that fit into that format.
func ELogMessagef(message string, args ...any) {
	if logFileAlive {
		formatted := fmt.Sprintf(message, args...)
		logFile.Write([]byte(time.Now().String() + ": " + formatted + "\n"))
	}
}

// Exclusively logs an error and its message to the log file.
// Params: The error, the message identifying that error
func ElogError(err error, message string) {
	if logFileAlive {
		logFile.Write([]byte("ERR at " + time.Now().String() + ": " + message + ": " + err.Error() + "\n"))
	}
}

// Logs a message to the console and log file, closes the log file, and crashes the server.
// Only to be used in setup, BEFORE the slack integration has been enabled.
func FatalLogMessage(message string) {
	LogMessage("FATAL: " + message)
	logFile.Close()
	os.Exit(1)
}

// Logs an error and its message to the console, log file, and slack before closign the log file and crashing the server.
func FatalError(err error, message string) {
	LogError(err, "FATAL: "+message)
	logFile.Close()
	NotifyOnline(false)
	os.Exit(1)
}

// A wrapper around filemanager.MkDirWithPermissions() that includes error handling.
func HandleMkdirAll(filepath string) {
	mkDirErr := filemanager.MkDirWithPermissions(filepath)

	if mkDirErr != nil {
		LogErrorf(mkDirErr, "Problem making directory %v", filepath)
	}
}

// Creates a new log.Logger that writes to the log file for passing into any
// Constructors that can take one, such as the http handler.
func GetLogger() *log.Logger {
	return log.New(
		logFile,
		"httplog: ",
		log.LstdFlags,
	)
}

// Shuts down the log file by closing the reference to it and setting logFileAlive to false
func ShutdownLogFile() {
	LogMessage("Shutting down log file due to configs...")
	logFile.Close()
	logFileAlive = false
}
