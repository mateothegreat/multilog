package multilog

import (
	"testing"
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
