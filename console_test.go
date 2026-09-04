package multilog

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestConsoleLogger_Handle(t *testing.T) {
	logger := NewConsoleLogger(&NewConsoleLoggerArgs{
		Format: FormatText,
		FilterDropPatterns: []*string{
			PtrString("block_this_group"),
			PtrString(".*drop.*"), // Drop any message that contains the word "drop"
		},
	})
	logger.Setup()
	logger.Log(INFO, "test", "test", map[string]interface{}{
		"foo": "test",
		"bar": 1,
	})
	logger.Log(WARN, "test", "test", map[string]interface{}{
		"foo": "test",
		"bar": 1,
	})
}

func TestFormatFieldsEmpty(t *testing.T) {
	if got := formatFields(nil, false); got != "" {
		t.Fatalf("nil map: got %q, want empty", got)
	}
	if got := formatFields(map[string]interface{}{}, true); got != "" {
		t.Fatalf("empty map: got %q, want empty", got)
	}
}

func TestFormatFieldsCompact(t *testing.T) {
	prev := color.NoColor
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = prev })

	got := formatFields(map[string]interface{}{
		"foo": "foo",
		"bar": 1,
		"baz": "baz",
	}, false)
	want := " bar: 1 baz: baz foo: foo"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFormatFieldsExpanded(t *testing.T) {
	prev := color.NoColor
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = prev })

	got := formatFields(map[string]interface{}{
		"foo": "foo",
		"bar": 1,
		"baz": "baz",
	}, true)
	want := "\n  bar: 1\n  baz: baz\n  foo: foo"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFormatFieldsColors(t *testing.T) {
	prev := color.NoColor
	color.NoColor = false
	t.Cleanup(func() { color.NoColor = prev })

	got := formatFields(map[string]interface{}{"foo": "bar"}, false)
	if !strings.Contains(got, fieldNameColor.Sprintf("%s:", "foo")) {
		t.Fatalf("key is not gray: %q", got)
	}
	if !strings.Contains(got, fieldValueColor.Sprintf("%v", "bar")) {
		t.Fatalf("value is not white: %q", got)
	}
	if strings.Contains(got, "map[") {
		t.Fatalf("still dumping a Go map: %q", got)
	}
}

func TestConsoleLoggerTextPrintsFieldsInline(t *testing.T) {
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

	logger := NewConsoleLogger(&NewConsoleLoggerArgs{
		Format: FormatText,
	})
	logger.Setup()
	logger.Log(DEBUG, "my_package_name", "test", map[string]interface{}{
		"foo": "foo",
		"bar": 1,
		"baz": "baz",
	})

	got := buf.String()
	want := "[DEBUG] my_package_name: test bar: 1 baz: baz foo: foo\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestConsoleLoggerTextExpandPrintsFieldsOnNewLines(t *testing.T) {
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

	logger := NewConsoleLogger(&NewConsoleLoggerArgs{
		Format: FormatText,
	})
	logger.Setup()
	logger.LogWith(DEBUG, "my_package_name", "test", map[string]interface{}{
		"foo": "foo",
		"bar": 1,
		"baz": "baz",
	}, Options{Expand: true})

	got := buf.String()
	want := "[DEBUG] my_package_name: test\n  bar: 1\n  baz: baz\n  foo: foo\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestConsoleLoggerTextOmitsFieldsWhenEmpty(t *testing.T) {
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

	logger := NewConsoleLogger(&NewConsoleLoggerArgs{
		Format: FormatText,
	})
	logger.Setup()
	logger.Log(INFO, "grp", "hello", nil)

	got := buf.String()
	want := "[INFO] grp: hello\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
