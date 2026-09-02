package multilog

import (
	"fmt"
	"sync"
)

// Loggers is a map of log methods to custom loggers.
//
// This is a global variable that is used to store the loggers for each log method.
// It is not recommended to use this variable directly, but rather to use the functions
// in the log package to register and retrieve loggers.
var Loggers map[LogMethod]*CustomLogger = make(map[LogMethod]*CustomLogger)

// loggersMu guards concurrent access to the Loggers map.
var loggersMu sync.RWMutex

// NewLogger creates a new logger for the given log method.
//
// Arguments:
//   - t: The log method to create a logger for.
//
// Returns:
//
//   - A new logger for the given log method.
//     If a logger for the given log method is already registered, it is returned.
//     Otherwise, a new logger is created and registered for the given log method.
func NewLogger(t LogMethod) *CustomLogger {
	loggersMu.Lock()
	defer loggersMu.Unlock()
	if existing, ok := Loggers[t]; ok {
		return existing
	}
	Loggers[t] = &CustomLogger{}
	return Loggers[t]
}

// ResetLoggers removes all registered loggers.
//
// Intended for use in tests to reset global state between cases.
// It must not be called concurrently with log calls.
func ResetLoggers() {
	loggersMu.Lock()
	defer loggersMu.Unlock()
	Loggers = make(map[LogMethod]*CustomLogger)
}

// RegisterLogger registers a custom logger for a given log method.
//
// Register loggers only from the main package's init() or main(). Library
// packages must not register loggers in their own init() functions, as import
// order is not guaranteed and duplicate registrations return an error.
//
// Arguments:
//   - t: The log method to register a logger for.
//   - logger: The custom logger to register.
//
// Returns:
//   - `error` if the logger for the given log method is already registered.
//     The duplicate check happens before logger.Setup is run, so a failed
//     registration has no side effects.
//   - `nil` if the logger for the given log method is successfully registered.
//     The logger is inserted into the Loggers map before logger.Setup runs,
//     so a Setup that logs recursively will find its own logger registered.
func RegisterLogger(t LogMethod, logger *CustomLogger) error {
	loggersMu.Lock()
	if _, exists := Loggers[t]; exists {
		loggersMu.Unlock()
		return fmt.Errorf("logger for log method %s already registered", t)
	}
	Loggers[t] = logger
	loggersMu.Unlock()

	if logger.Setup != nil {
		logger.Setup()
	}
	return nil
}
