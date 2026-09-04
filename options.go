package multilog

// Options holds per-call configuration applied by With.
//
// Zero value is the default: fields stay on one line. Set Expand to print each
// field on its own line. Add new fields here as you introduce more Option
// helpers so With stays a single variadic entry point.
type Options struct {
	// Expand prints each field on its own line when true. When false, fields
	// stay on the same line as the message.
	Expand bool
}

// Option is a single configuration item for With. Pass any number of them:
//
//	multilog.With(multilog.Expand).Info("api", "ready", data)
type Option func(*Options)

// Expand is an Option that prints each structured field on its own line.
//
// Without it, console text output keeps fields on the same line as the
// message. JSON loggers and custom loggers that do not implement LogWith
// ignore this option.
func Expand(o *Options) {
	o.Expand = true
}

// Scope carries options for subsequent log calls. Create one with With.
//
// You can reuse a Scope across calls:
//
//	expanded := multilog.With(multilog.Expand)
//	expanded.Info("api", "ready", data)
//	expanded.Error("api", "failed", errData)
//
// Prefer With(...).Info(...) over Info(...).With(...). Package-level Info
// logs immediately, so a trailing With would run too late to change the line.
type Scope struct {
	opts Options
}

// With returns a Scope that applies opts to every log method you call on it.
//
//	multilog.With(multilog.Expand).Info("api", "ready", data)
//
// Package-level Trace/Debug/Info/Warn/Error/Fatal keep the default (compact)
// layout. Nil options are skipped so you can pass optional flags safely.
func With(opts ...Option) Scope {
	var o Options
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return Scope{opts: o}
}

// mergeOptions combines logger defaults with per-call options. A flag is on if
// either side set it, so a logger registered with Expand stays expanded even
// when the caller uses the package-level Info functions.
func mergeOptions(base Options, call Options) Options {
	return Options{
		Expand: base.Expand || call.Expand,
	}
}

// Trace logs at TRACE using the options from With.
func (s Scope) Trace(group string, message string, v map[string]interface{}) {
	emit(TRACE, group, message, v, s.opts)
}

// Debug logs at DEBUG using the options from With.
func (s Scope) Debug(group string, message string, v map[string]interface{}) {
	emit(DEBUG, group, message, v, s.opts)
}

// Info logs at INFO using the options from With.
func (s Scope) Info(group string, message string, v map[string]interface{}) {
	emit(INFO, group, message, v, s.opts)
}

// Warn logs at WARN using the options from With.
func (s Scope) Warn(group string, message string, v map[string]interface{}) {
	emit(WARN, group, message, v, s.opts)
}

// Error logs at ERROR using the options from With.
func (s Scope) Error(group string, message string, v map[string]interface{}) {
	emit(ERROR, group, message, v, s.opts)
}

// Fatal logs at FATAL using the options from With, then exits the process.
func (s Scope) Fatal(group string, message string, v map[string]interface{}) {
	emit(FATAL, group, message, v, s.opts)
	finishFatal()
}
