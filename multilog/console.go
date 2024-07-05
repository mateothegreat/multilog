package multilog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

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
	args   *NewConsoleLoggerArgs
	logger *slog.Logger // logger is the slog.Logger instance used for logging.
}

// Setup initializes the CustomLogger by creating a new slog.Logger.
func (c *ConsoleLogger) Setup() {
	c.logger = NewSlogLogger()
}

// Log logs a message with the given log level, group, message, and additional data.
func (c *ConsoleLogger) Log(level LogLevel, group string, message string, v any) {
	logger := c.logger.With(slog.String("group", group))
	switch level {
	case DEBUG:
		if c.args.Format == FormatJSON {
			logger.Debug(message, "data", v)
		} else {
			log.Printf(color.CyanString("[DEBUG]")+" %s: %s %v", color.GreenString(group), color.YellowString(message), v)
		}
	case INFO:
		if c.args.Format == FormatJSON {
			logger.Info(message, "data", v)
		} else {
			log.Printf(color.BlueString("[INFO]")+" %s: %s %v", color.GreenString(group), color.YellowString(message), v)
		}
	case WARN:
		if c.args.Format == FormatJSON {
			logger.Warn(message, "data", v)
		} else {
			log.Printf(color.YellowString("[WARN]")+" %s: %s %v", color.GreenString(group), color.YellowString(message), v)
		}
	case ERROR:
		if c.args.Format == FormatJSON {
			logger.Error(message, "data", v)
		} else {
			log.Printf(color.RedString("[ERROR]")+" %s: %s %v", color.GreenString(group), color.YellowString(message), v)
		}
	default:
		if c.args.Format == FormatJSON {
			logger.Info(message, "data", v)
		} else {
			log.Printf(color.BlueString("[INFO]")+" %s: %s %v", color.GreenString(group), color.YellowString(message), v)
		}
	}
}

// Format is the format of the log that is output.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

type NewConsoleLoggerArgs struct {
	Format Format
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
