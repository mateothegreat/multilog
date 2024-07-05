package logging

import (
	"sync"

	"github.com/mateothegreat/go-multilog/logging/types"
)

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
	for _, logger := range types.Loggers {
		wg.Add(1)
		go func(logger *types.CustomLogger) {
			defer wg.Done()
			logger.Log(types.DEBUG, group, message, v)
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
	for _, logger := range types.Loggers {
		wg.Add(1)
		go func(logger *types.CustomLogger) {
			defer wg.Done()
			logger.Log(types.INFO, group, message, v)
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
	for _, logger := range types.Loggers {
		wg.Add(1)
		go func(logger *types.CustomLogger) {
			defer wg.Done()
			logger.Log(types.WARN, group, message, v)
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
	for _, logger := range types.Loggers {
		wg.Add(1)
		go func(logger *types.CustomLogger) {
			defer wg.Done()
			logger.Log(types.ERROR, group, message, v)
		}(logger)
	}
	wg.Wait()
}
