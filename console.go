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
	"sort"
	"strings"

	"github.com/fatih/color"
)

// fieldNameColor is a true mid-gray. color.HiBlack (ANSI 90) is "bright black"
// and most themes map it to the default foreground or near-black, so keys never
// look gray.
var fieldNameColor = color.RGB(158, 158, 158)

// fieldValueColor is bright white so values stay distinct from gray keys.
var fieldValueColor = color.New(color.FgHiWhite)

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
// Fields stay on one line. Use LogWith with Expand for a multiline layout.
func (c *ConsoleLogger) Log(level LogLevel, group string, message string, v map[string]interface{}) {
	c.log(level, group, message, v, Options{})
}

// LogWith logs a message and honors per-call options from With, such as Expand.
func (c *ConsoleLogger) LogWith(level LogLevel, group string, message string, v map[string]interface{}, opts Options) {
	c.log(level, group, message, v, opts)
}

// log is the shared console write path for Log and LogWith.
func (c *ConsoleLogger) log(level LogLevel, group string, message string, v map[string]interface{}, opts Options) {
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
	case TRACE:
		if c.args.Format == FormatJSON {
			// slog has no native TRACE level below Debug; use Debug instead.
			logger.Debug(message, "data", v)
		} else {
			logText(color.HiMagentaString("[TRACE]"), group, message, v, opts.Expand)
		}
	case DEBUG:
		if c.args.Format == FormatJSON {
			logger.Debug(message, "data", v)
		} else {
			logText(color.HiCyanString("[DEBUG]"), group, message, v, opts.Expand)
		}
	case INFO:
		if c.args.Format == FormatJSON {
			logger.Info(message, "data", v)
		} else {
			logText(color.HiBlueString("[INFO]"), group, message, v, opts.Expand)
		}
	case WARN:
		if c.args.Format == FormatJSON {
			logger.Warn(message, "data", v)
		} else {
			logText(color.HiYellowString("[WARN]"), group, message, v, opts.Expand)
		}
	case ERROR:
		if c.args.Format == FormatJSON {
			logger.Error(message, "data", v)
		} else {
			logText(color.HiRedString("[ERROR]"), group, message, v, opts.Expand)
		}
	case FATAL:
		if c.args.Format == FormatJSON {
			logger.Error(message, "data", v)
		} else {
			logText(color.HiRedString("[FATAL]"), group, message, v, opts.Expand)
		}
	default:
		if c.args.Format == FormatJSON {
			logger.Info(message, "data", v)
		} else {
			logText(color.HiBlueString("[UNKNOWN]"), group, message, v, opts.Expand)
		}
	}
}

// logText writes a text-format console line: colored level, group, message, then
// the fields from v. Expand puts each field on its own line; otherwise they stay
// on the same line as the message.
func logText(level string, group string, message string, v map[string]interface{}, expand bool) {
	log.Printf("%s %s: %s%s", level, color.GreenString(group), color.YellowString(message), formatFields(v, expand))
}

// formatFields renders v with gray keys and white values. Keys are sorted so the
// same payload always prints in the same order. A nil or empty map produces no
// extra output. When expand is true, each field is on its own indented line.
func formatFields(v map[string]interface{}, expand bool) string {
	if len(v) == 0 {
		return ""
	}

	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		if expand {
			b.WriteString("\n  ")
		} else {
			b.WriteByte(' ')
		}
		b.WriteString(fieldNameColor.Sprintf("%s:", k))
		b.WriteByte(' ')
		b.WriteString(fieldValueColor.Sprintf("%v", v[k]))
	}
	return b.String()
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
		Setup:   logger.Setup,
		Log:     logger.Log,
		LogWith: logger.LogWith,
	}
}
