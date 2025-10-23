package multilog

import "fmt"

// Loggers is a map of log methods to custom loggers.
//
// This is a global variable that is used to store the loggers for each log method.
// It is not recommended to use this variable directly, but rather to use the functions
// in the log package to register and retrieve loggers.
var Loggers map[LogMethod]*CustomLogger = make(map[LogMethod]*CustomLogger)

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
	Loggers[t] = &CustomLogger{}
	return Loggers[t]
}

// RegisterLogger registers a custom logger for a given log method.
//
// Arguments:
//   - t: The log method to register a logger for.
//   - logger: The custom logger to register.
//
// Returns:
//   - `error` if the logger for the given log method is already registered.
//   - `nil` if the logger for the given log method is successfully registered.
func RegisterLogger(t LogMethod, logger *CustomLogger) error {
	if _, exists := Loggers[t]; exists {
		return fmt.Errorf("logger for log method %s already registered", t)
	}

	if logger.Setup != nil {
		logger.Setup()
	}

	Loggers[t] = logger
	return nil
}
