package multilog

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestWithAppliesExpand(t *testing.T) {
	scope := With(Expand)
	if !scope.opts.Expand {
		t.Fatal("With(Expand) did not set Expand")
	}
}

func TestWithNoOptionsStaysCompact(t *testing.T) {
	scope := With()
	if scope.opts.Expand {
		t.Fatal("With() should leave Expand false")
	}
}

func TestWithSkipsNilOptions(t *testing.T) {
	scope := With(nil, Expand, nil)
	if !scope.opts.Expand {
		t.Fatal("With(nil, Expand, nil) should still apply Expand")
	}
}

func TestWithExpandLogsMultilineThroughRegistry(t *testing.T) {
	prevColor := color.NoColor
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = prevColor })

	var buf bytes.Buffer
	prevOut := log.Writer()
	prevFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	})

	ResetLoggers()
	t.Cleanup(ResetLoggers)
	if err := RegisterLogger(LogMethod("console"), NewConsoleLogger(&NewConsoleLoggerArgs{
		Format: FormatText,
	})); err != nil {
		t.Fatalf("register: %v", err)
	}

	payload := map[string]interface{}{"foo": "foo", "bar": 1}

	Info("grp", "compact", payload)
	if strings.Contains(buf.String(), "\n  ") {
		t.Fatalf("package Info should stay on one line: %q", buf.String())
	}

	buf.Reset()
	With(Expand).Info("grp", "expanded", payload)
	got := buf.String()
	want := "[INFO] grp: expanded\n  bar: 1\n  foo: foo\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestScopeAllLevelsHonorExpand(t *testing.T) {
	prevColor := color.NoColor
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = prevColor })

	var buf bytes.Buffer
	prevOut := log.Writer()
	prevFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	})

	ResetLoggers()
	t.Cleanup(ResetLoggers)
	if err := RegisterLogger(LogMethod("console"), NewConsoleLogger(&NewConsoleLoggerArgs{
		Level:  TRACE,
		Format: FormatText,
	})); err != nil {
		t.Fatalf("register: %v", err)
	}

	payload := map[string]interface{}{"k": "v"}
	expanded := With(Expand)

	cases := []struct {
		name string
		log  func()
		want string
	}{
		{"trace", func() { expanded.Trace("g", "m", payload) }, "[TRACE] g: m\n  k: v\n"},
		{"debug", func() { expanded.Debug("g", "m", payload) }, "[DEBUG] g: m\n  k: v\n"},
		{"info", func() { expanded.Info("g", "m", payload) }, "[INFO] g: m\n  k: v\n"},
		{"warn", func() { expanded.Warn("g", "m", payload) }, "[WARN] g: m\n  k: v\n"},
		{"error", func() { expanded.Error("g", "m", payload) }, "[ERROR] g: m\n  k: v\n"},
	}

	for _, tc := range cases {
		buf.Reset()
		tc.log()
		if buf.String() != tc.want {
			t.Fatalf("%s: got %q, want %q", tc.name, buf.String(), tc.want)
		}
	}
}

func TestScopeFatalHonorsExpand(t *testing.T) {
	prevColor := color.NoColor
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = prevColor })

	var buf bytes.Buffer
	prevOut := log.Writer()
	prevFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	})

	prevExit := exitFn
	prevFlush := fatalFlush
	var exited int
	exitFn = func(code int) { exited = code }
	fatalFlush = func() {}
	t.Cleanup(func() {
		exitFn = prevExit
		fatalFlush = prevFlush
	})

	ResetLoggers()
	t.Cleanup(ResetLoggers)
	if err := RegisterLogger(LogMethod("console"), NewConsoleLogger(&NewConsoleLoggerArgs{
		Format: FormatText,
	})); err != nil {
		t.Fatalf("register: %v", err)
	}

	With(Expand).Fatal("g", "m", map[string]interface{}{"k": "v"})

	want := "[FATAL] g: m\n  k: v\n"
	if buf.String() != want {
		t.Fatalf("got %q, want %q", buf.String(), want)
	}
	if exited != 1 {
		t.Fatalf("exit code: got %d, want 1", exited)
	}
}

func TestRegisterLoggerWithExpandAppliesToAllLevels(t *testing.T) {
	prevColor := color.NoColor
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = prevColor })

	var buf bytes.Buffer
	prevOut := log.Writer()
	prevFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	})

	ResetLoggers()
	t.Cleanup(ResetLoggers)
	if err := RegisterLogger(LogMethod("console"), NewConsoleLogger(&NewConsoleLoggerArgs{
		Level:  TRACE,
		Format: FormatText,
	}), With(Expand)); err != nil {
		t.Fatalf("register: %v", err)
	}

	payload := map[string]interface{}{"k": "v"}
	cases := []struct {
		name string
		log  func()
		want string
	}{
		{"trace", func() { Trace("g", "m", payload) }, "[TRACE] g: m\n  k: v\n"},
		{"debug", func() { Debug("g", "m", payload) }, "[DEBUG] g: m\n  k: v\n"},
		{"info", func() { Info("g", "m", payload) }, "[INFO] g: m\n  k: v\n"},
		{"warn", func() { Warn("g", "m", payload) }, "[WARN] g: m\n  k: v\n"},
		{"error", func() { Error("g", "m", payload) }, "[ERROR] g: m\n  k: v\n"},
	}

	for _, tc := range cases {
		buf.Reset()
		tc.log()
		if buf.String() != tc.want {
			t.Fatalf("%s: got %q, want %q", tc.name, buf.String(), tc.want)
		}
	}
}

func TestRegisterLoggerStoresMergedOptions(t *testing.T) {
	ResetLoggers()
	t.Cleanup(ResetLoggers)

	logger := NewConsoleLogger(&NewConsoleLoggerArgs{Format: FormatText})
	if err := RegisterLogger(LogMethod("console"), logger, With(Expand)); err != nil {
		t.Fatalf("register: %v", err)
	}
	if !logger.Options.Expand {
		t.Fatal("RegisterLogger(..., With(Expand)) did not set logger.Options.Expand")
	}
}

func TestEmitFallsBackToLogWhenLogWithIsNil(t *testing.T) {
	ResetLoggers()
	t.Cleanup(ResetLoggers)

	var called bool
	if err := RegisterLogger(LogMethod("custom"), &CustomLogger{
		Log: func(level LogLevel, group string, message string, v map[string]interface{}) {
			called = true
			if group != "g" || message != "m" {
				t.Fatalf("unexpected args: %s %s", group, message)
			}
		},
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	With(Expand).Info("g", "m", nil)
	if !called {
		t.Fatal("custom logger Log was not called")
	}
}
