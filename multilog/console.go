package multilog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"regexp"

	"github.com/fatih/color"
)

// PrettyHandlerOptions defines options for the PrettyHandler.
type PrettyHandlerOptions struct {
	SlogOpts slog.HandlerOptions // SlogOpts are the options for the slog.Handler.
}

// PrettyHandler is a custom handler for pretty-printing log messages.
type PrettyHandler struct {
	slog.Handler
	l *log.Logger // l is the standard library logger used for output.
}

// Handle processes the log record and outputs it in a pretty format.
func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	level := fmt.Sprintf("[%s]", r.Level.String())

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})

	b, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}

	timeStr := r.Time.Format("[15:05:05.000]")
	msg := color.CyanString(r.Message)

	h.l.Println(timeStr, level, msg, color.WhiteString(string(b)))

	return nil
}

// NewPrettyHandler creates a new PrettyHandler with the given output writer and options.
func NewPrettyHandler(out io.Writer, opts PrettyHandlerOptions) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, &opts.SlogOpts),
		l:       log.New(out, "", 0),
	}

	return h
}

// NewSlogLogger creates a new slog.Logger with a PrettyHandler.
func NewSlogLogger() *slog.Logger {
	return slog.New(NewPrettyHandler(os.Stdout, PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}))
}

// ConsoleLogger is a custom logger that uses slog.Logger.
type ConsoleLogger struct {
	args           *NewConsoleLoggerArgs // args are the arguments for the NewConsoleLogger function.
	logger         *slog.Logger          // logger is the slog.Logger instance used for logging.
	filterPatterns []*regexp.Regexp      // filterPatterns are the regex patterns to filter out log messages.
}

// Setup initializes the CustomLogger by creating a new slog.Logger.
func (c *ConsoleLogger) Setup() {
	// Create a new slog.Logger.
	c.logger = NewSlogLogger()
	// Compile the filter drop patterns into regexp.Regexp instances.
	for _, pattern := range c.args.FilterDropPatterns {
		if pattern != nil {
			c.filterPatterns = append(c.filterPatterns, regexp.MustCompile(*pattern))
		}
	}
}

// Log logs a message with the given log level, group, message, and additional data.
func (c *ConsoleLogger) Log(level LogLevel, group string, message string, v map[string]interface{}) {
	// Check if the log level is sufficient to log the message.
	if level < c.args.Level {
		return // Drop the message if the log level is lower than the configured level.
	}

	// Check if the message matches any of the filter drop patterns.
	for _, pattern := range c.filterPatterns {
		if pattern.MatchString(group) || pattern.MatchString(message) {
			return
		}
	}

	// Create a new slog.Logger with the group.
	logger := c.logger.With(slog.String("group", group))

	// Log the message with the given log level.
	switch level {
	case DEBUG:
		if c.args.Format == FormatJSON {
			logger.Debug(message, "data", v)
		} else {
			log.Printf(color.HiCyanString("[DEBUG]")+" %s: %s %v", color.GreenString(group), color.YellowString(message), colorizeMap(v))
		}
	case INFO:
		if c.args.Format == FormatJSON {
			logger.Info(message, "data", v)
		} else {
			log.Printf(color.HiBlueString("[INFO]")+" %s: %s %v", color.GreenString(group), color.YellowString(message), colorizeMap(v))
		}
	case WARN:
		if c.args.Format == FormatJSON {
			logger.Warn(message, "data", v)
		} else {
			log.Printf(color.HiYellowString("[WARN]")+" %s: %s %v", color.GreenString(group), color.YellowString(message), colorizeMap(v))
		}
	case ERROR:
		if c.args.Format == FormatJSON {
			logger.Error(message, "data", v)
		} else {
			log.Printf(color.HiRedString("[ERROR]")+" %s: %s %v", color.GreenString(group), color.YellowString(message), colorizeMap(v))
		}
	default:
		if c.args.Format == FormatJSON {
			logger.Info(message, "data", v)
		} else {
			log.Printf(color.HiBlueString("[INFO]")+" %s: %s %v", color.GreenString(group), color.YellowString(message), colorizeMap(v))
		}
	}
}

func colorizeMap(v map[string]interface{}) map[string]interface{} {
	colorizedMap := make(map[string]interface{})
	for key, value := range v {
		colorizedMap[color.HiBlueString(key)] = color.HiBlackString("%v", value)
	}
	return colorizedMap
}

// Format is the format of the log that is output.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// NewConsoleLoggerArgs are the arguments for the NewConsoleLogger function.
type NewConsoleLoggerArgs struct {
	// Level is the log level to use.
	Level LogLevel
	// Format is the format of the log that is output.
	Format Format
	// FilterDropPatterns is a slice of regex patterns to filter out log messages.
	FilterDropPatterns []*string
}

// NewConsoleLogger creates a new CustomLogger for console logging.
//
// Returns a new CustomLogger with the setup and log functions for console logging.
func NewConsoleLogger(args *NewConsoleLoggerArgs) *CustomLogger {
	logger := &ConsoleLogger{
		args: args,
	}

	return &CustomLogger{
		Setup: logger.Setup,
		Log:   logger.Log,
	}
}
