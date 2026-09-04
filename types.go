package multilog

// LogFn is a function type that defines the signature for logging functions.
// It takes a log level, group name, message, and additional data as arguments.
type LogFn func(level LogLevel, group string, message string, v map[string]interface{})

// LogMethod represents the method used for logging, such as console or elasticsearch.
type LogMethod string

// LogLevel represents the severity level of a log message.
type LogLevel int

// CustomLogger is a struct that defines a custom logger with setup and log functions.
type CustomLogger struct {
	Setup func() // Setup is a function that initializes the custom logger.
	Log   LogFn  // Log is a function that logs a message with a given log level, group, message, and additional data.
	// LogWith is an optional function that receives per-call options from With.
	// When it is set, emit prefers it over Log so options like Expand are honored.
	// Custom loggers may leave it nil; they keep receiving calls through Log.
	LogWith LogWithFn
	// Options are defaults for this logger. RegisterLogger(..., With(Expand))
	// sets them so every call expands without a per-call With. emit ORs these
	// with the call's options before invoking LogWith.
	Options Options
}

// LogWithFn is the signature for loggers that accept per-call options from With.
type LogWithFn func(level LogLevel, group string, message string, v map[string]interface{}, opts Options)

// Logger is an interface that defines the methods required for a logger.
type Logger interface {
	Setup() // Setup initializes the logger.
	// Log logs a message with a given log level, group, message, and additional data.
	Log(level LogLevel, group string, message string, v map[string]interface{})
}

const (
	// TRACE represents the trace log level.
	TRACE LogLevel = LogLevel(0)
	// DEBUG represents the debug log level.
	DEBUG LogLevel = LogLevel(1)
	// INFO represents the info log level.
	INFO LogLevel = LogLevel(2)
	// WARN represents the warn log level.
	WARN LogLevel = LogLevel(3)
	// ERROR represents the error log level.
	ERROR LogLevel = LogLevel(4)
	// FATAL represents the fatal log level.
	FATAL LogLevel = LogLevel(5)
)

const (
	// LoggerConsole represents the console log method.
	LoggerConsole LogMethod = "console"
	// LoggerElasticsearch represents the elasticsearch log method.
	LoggerElasticsearch LogMethod = "elasticsearch"
)
