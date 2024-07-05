package multilog

// LogFn is a function type that defines the signature for logging functions.
// It takes a log level, group name, message, and additional data as arguments.
type LogFn func(level LogLevel, group string, message string, v any)

// LogMethod represents the method used for logging, such as console or elasticsearch.
type LogMethod string

// LogLevel represents the severity level of a log message.
type LogLevel int

// CustomLogger is a struct that defines a custom logger with setup and log functions.
type CustomLogger struct {
	Setup func() // Setup is a function that initializes the custom logger.
	Log   LogFn  // Log is a function that logs a message with a given log level, group, message, and additional data.
}

// Logger is an interface that defines the methods required for a logger.
type Logger interface {
	Setup()                                                  // Setup initializes the logger.
	Log(level LogLevel, group string, message string, v any) // Log logs a message with a given log level, group, message, and additional data.
}

const (
	// DEBUG represents the debug log level.
	DEBUG LogLevel = LogLevel(0)
	// INFO represents the info log level.
	INFO LogLevel = LogLevel(1)
	// WARN represents the warn log level.
	WARN LogLevel = LogLevel(2)
	// ERROR represents the error log level.
	ERROR LogLevel = LogLevel(3)
)

const (
	// LoggerConsole represents the console log method.
	LoggerConsole LogMethod = "console"
	// LoggerElasticsearch represents the elasticsearch log method.
	LoggerElasticsearch LogMethod = "elasticsearch"
)
