package multilog

import (
	"os"
	"sync"
	"time"
)

// snapshotLoggers returns a copy of the currently registered loggers.
//
// The read lock is held only long enough to copy the logger references into
// a local slice. Iteration happens after the lock is released so a logger
// that re-registers (or logs recursively) cannot deadlock against the map.
func snapshotLoggers() []*CustomLogger {
	loggersMu.RLock()
	defer loggersMu.RUnlock()
	loggers := make([]*CustomLogger, 0, len(Loggers))
	for _, logger := range Loggers {
		loggers = append(loggers, logger)
	}
	return loggers
}

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
func Trace(group string, message string, v map[string]interface{}) {
	wg := sync.WaitGroup{}
	for _, logger := range snapshotLoggers() {
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
func Debug(group string, message string, v map[string]interface{}) {
	wg := sync.WaitGroup{}
	for _, logger := range snapshotLoggers() {
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
func Info(group string, message string, v map[string]interface{}) {
	wg := sync.WaitGroup{}
	for _, logger := range snapshotLoggers() {
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
func Warn(group string, message string, v map[string]interface{}) {
	wg := sync.WaitGroup{}
	for _, logger := range snapshotLoggers() {
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
func Error(group string, message string, v map[string]interface{}) {
	wg := sync.WaitGroup{}
	for _, logger := range snapshotLoggers() {
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
func Fatal(group string, message string, v map[string]interface{}) {
	wg := sync.WaitGroup{}
	for _, logger := range snapshotLoggers() {
		wg.Add(1)
		go func(logger *CustomLogger) {
			defer wg.Done()
			logger.Log(FATAL, group, message, v)
		}(logger)
	}
	wg.Wait()

	// Best-effort flush: give buffered logger backends (e.g. the async
	// elasticsearch client) a brief window to drain their queues before the
	// process exits. This is NOT a guarantee of delivery — os.Exit skips
	// deferred functions and cannot wait on internal flush goroutines.
	time.Sleep(100 * time.Millisecond)

	os.Exit(1)
}
