package multilog

import (
	"os"
	"sync"
	"time"
)

// exitFn is called by Fatal after loggers complete. It is a variable so
// tests can stub it out. Not safe to mutate concurrently with Fatal calls.
var exitFn = os.Exit

// fatalFlush is the best-effort pause before os.Exit. Tests replace it so Fatal
// does not sleep. Not safe to mutate concurrently with Fatal calls.
var fatalFlush = func() { time.Sleep(100 * time.Millisecond) }

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

// emit fans a log call out to every registered logger and waits for them.
//
// Loggers that implement LogWith receive opts (including Expand). Everyone
// else is called through Log and ignores options. This keeps existing custom
// loggers working without a signature change.
func emit(level LogLevel, group string, message string, v map[string]interface{}, opts Options) {
	wg := sync.WaitGroup{}
	for _, logger := range snapshotLoggers() {
		wg.Add(1)
		go func(logger *CustomLogger) {
			defer wg.Done()
			if logger.LogWith != nil {
				logger.LogWith(level, group, message, v, opts)
				return
			}
			if logger.Log != nil {
				logger.Log(level, group, message, v)
			}
		}(logger)
	}
	wg.Wait()
}

// finishFatal flushes buffered backends, then exits. Shared by Fatal and
// Scope.Fatal so both paths behave the same.
func finishFatal() {
	// Best-effort flush: give buffered logger backends (e.g. the async
	// elasticsearch client) a brief window to drain their queues before the
	// process exits. This is NOT a guarantee of delivery — os.Exit skips
	// deferred functions and cannot wait on internal flush goroutines.
	fatalFlush()
	exitFn(1)
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
//   - v: The data to log (must not be mutated after the call; it is shared
//     across logger goroutines)
func Trace(group string, message string, v map[string]interface{}) {
	emit(TRACE, group, message, v, Options{})
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
//   - v: The data to log (must not be mutated after the call; it is shared
//     across logger goroutines)
func Debug(group string, message string, v map[string]interface{}) {
	emit(DEBUG, group, message, v, Options{})
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
//   - v: The data to log (must not be mutated after the call; it is shared
//     across logger goroutines)
func Info(group string, message string, v map[string]interface{}) {
	emit(INFO, group, message, v, Options{})
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
//   - v: The data to log (must not be mutated after the call; it is shared
//     across logger goroutines)
func Warn(group string, message string, v map[string]interface{}) {
	emit(WARN, group, message, v, Options{})
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
//   - v: The data to log (must not be mutated after the call; it is shared
//     across logger goroutines)
func Error(group string, message string, v map[string]interface{}) {
	emit(ERROR, group, message, v, Options{})
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
//   - v: The data to log (must not be mutated after the call; it is shared
//     across logger goroutines)
func Fatal(group string, message string, v map[string]interface{}) {
	emit(FATAL, group, message, v, Options{})
	finishFatal()
}
