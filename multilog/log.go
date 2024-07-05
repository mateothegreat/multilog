package multilog

import (
	"os"
	"sync"
)

// Trace logs a trace message to all registered loggers at the TRACE level.
//
// This function is concurrently called for each logger, so it is safe to call
// from multiple goroutines without blocking.
//
// Arguments:
//
//   - group: The group name
//   - message: The message to log
//   - v: The data to log
func Trace(group string, message string, v any) {
	wg := sync.WaitGroup{}
	for _, logger := range Loggers {
		wg.Add(1)
		go func(logger *CustomLogger) {
			defer wg.Done()
			logger.Log(TRACE, group, message, v)
		}(logger)
	}
	wg.Wait()
}

// Debug logs a debug message to all registered loggers at the DEBUG level.
//
// This function is concurrently called for each logger, so it is safe to call
// from multiple goroutines without blocking.
//
// Arguments:
//
//   - group: The group name
//   - message: The message to log
//   - v: The data to log
func Debug(group string, message string, v any) {
	wg := sync.WaitGroup{}
	for _, logger := range Loggers {
		wg.Add(1)
		go func(logger *CustomLogger) {
			defer wg.Done()
			logger.Log(DEBUG, group, message, v)
		}(logger)
	}
	wg.Wait()
}

// Info logs an info message to all registered loggers at the INFO level.
//
// This function is concurrently called for each logger, so it is safe to call
// from multiple goroutines without blocking.
//
// Arguments:
//
//   - group: The group name
//   - message: The message to log
//   - v: The data to log
func Info(group string, message string, v any) {
	wg := sync.WaitGroup{}
	for _, logger := range Loggers {
		wg.Add(1)
		go func(logger *CustomLogger) {
			defer wg.Done()
			logger.Log(INFO, group, message, v)
		}(logger)
	}
	wg.Wait()
}

// Warn logs a warn message to all registered loggers at the WARN level.
//
// This function is concurrently called for each logger, so it is safe to call
// from multiple goroutines without blocking.
//
// Arguments:
//
//   - group: The group name
//   - message: The message to log
//   - v: The data to log
func Warn(group string, message string, v any) {
	wg := sync.WaitGroup{}
	for _, logger := range Loggers {
		wg.Add(1)
		go func(logger *CustomLogger) {
			defer wg.Done()
			logger.Log(WARN, group, message, v)
		}(logger)
	}
	wg.Wait()
}

// Error logs an error message to all registered loggers at the ERROR level.
//
// This function is concurrently called for each logger, so it is safe to call
// from multiple goroutines without blocking.
//
// Arguments:
//
//   - group: The group name
//   - message: The message to log
//   - v: The data to log
func Error(group string, message string, v any) {
	wg := sync.WaitGroup{}
	for _, logger := range Loggers {
		wg.Add(1)
		go func(logger *CustomLogger) {
			defer wg.Done()
			logger.Log(ERROR, group, message, v)
		}(logger)
	}
	wg.Wait()
}

// Fatal logs a fatal message to all registered loggers at the FATAL level.
//
// This function is concurrently called for each logger, so it is safe to call
// from multiple goroutines without blocking.
//
// Arguments:
//
//   - group: The group name
//   - message: The message to log
//   - v: The data to log
func Fatal(group string, message string, v any) {
	wg := sync.WaitGroup{}
	for _, logger := range Loggers {
		wg.Add(1)
		go func(logger *CustomLogger) {
			defer wg.Done()
			logger.Log(FATAL, group, message, v)
		}(logger)
	}
	wg.Wait()
	os.Exit(1)
}
