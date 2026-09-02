package multilog

import (
	"fmt"
	"sync/atomic"
	"testing"
)

// benchSink keeps no-op logger calls from being eliminated by the compiler.
var benchSink atomic.Int64

// benchSnapshot holds the result of snapshotLoggers so the call is not elided.
var benchSnapshot []*CustomLogger

// noopLogFn is a logger that performs no I/O and allocates nothing in the hot
// path; it only bumps a global counter so the call cannot be optimized away.
func noopLogFn(level LogLevel, group string, message string, v map[string]interface{}) {
	benchSink.Add(1)
}

// registerBenchLoggers registers n no-op loggers and arranges for the global
// registry to be reset when the benchmark finishes. Because Loggers is global,
// registration state must be created per benchmark and never shared across
// parallel sub-benchmarks.
func registerBenchLoggers(b *testing.B, n int) {
	b.Helper()
	for i := 0; i < n; i++ {
		err := RegisterLogger(LogMethod(fmt.Sprintf("bench-%d", i)), &CustomLogger{
			Log: noopLogFn,
		})
		if err != nil {
			b.Fatalf("register no-op logger %d: %v", i, err)
		}
	}
	b.Cleanup(ResetLoggers)
}

// BenchmarkInfo_NoLoggers measures the snapshot and empty-loop overhead of an
// Info call when no loggers are registered.
func BenchmarkInfo_NoLoggers(b *testing.B) {
	b.ReportAllocs()
	ResetLoggers()
	b.Cleanup(ResetLoggers)
	b.ResetTimer()
	for b.Loop() {
		Info("bench", "no loggers", nil)
	}
}

// BenchmarkInfo_SingleLogger measures sequential Info calls fanning out to a
// single no-op logger (one goroutine plus WaitGroup per call).
func BenchmarkInfo_SingleLogger(b *testing.B) {
	registerBenchLoggers(b, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		Info("bench", "single logger", nil)
	}
}

// BenchmarkInfo_MultiLogger measures fan-out scaling across 2, 4, and 8
// registered no-op loggers.
func BenchmarkInfo_MultiLogger(b *testing.B) {
	for _, n := range []int{2, 4, 8} {
		b.Run(fmt.Sprintf("loggers=%d", n), func(b *testing.B) {
			registerBenchLoggers(b, n)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				Info("bench", "multi logger", nil)
			}
		})
	}
}

// BenchmarkInfo_Concurrent measures throughput of Info under concurrent load:
// many goroutines fanning out to 4 registered no-op loggers.
func BenchmarkInfo_Concurrent(b *testing.B) {
	registerBenchLoggers(b, 4)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Info("bench", "concurrent", nil)
		}
	})
}

// BenchmarkSnapshotLoggers directly measures the RLock + copy path of
// snapshotLoggers with 8 registered loggers.
func BenchmarkSnapshotLoggers(b *testing.B) {
	registerBenchLoggers(b, 8)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		benchSnapshot = snapshotLoggers()
	}
}

// BenchmarkInfo_WithData measures the allocation cost of a typical Info call
// carrying a realistic structured payload (string, int, bool, float, nested
// map) to a single no-op logger.
func BenchmarkInfo_WithData(b *testing.B) {
	registerBenchLoggers(b, 1)
	b.ReportAllocs()
	b.ResetTimer()
	i := 0
	for b.Loop() {
		Info("bench", "with data", map[string]interface{}{
			"user":    "bench-user",
			"attempt": i,
			"success": true,
			"latency": 1.5,
			"nested": map[string]interface{}{
				"key": "value",
			},
		})
		i++
	}
}
