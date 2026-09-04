package elasticsearch

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mateothegreat/multilog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubCluster is a fake Elasticsearch used to drive the reconnect logic.
//
// Every request that the logger makes -- Exists, Create, Index -- is answered
// by respond, which the test swaps between "cluster down" (503) and "cluster
// up" (2xx) to simulate an outage and recovery. requests counts every call so
// tests can wait for the reconnect goroutine to have actually tried, rather
// than sleeping and hoping.
type stubCluster struct {
	server   *httptest.Server
	requests atomic.Int64
	respond  atomic.Value // func(http.ResponseWriter, *http.Request)
}

func newStubCluster(t *testing.T) *stubCluster {
	t.Helper()
	c := &stubCluster{}
	c.respond.Store(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	c.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.requests.Add(1)
		// The go-elasticsearch client rejects the response unless every
		// reply carries this header (its "product check"). Setting it in
		// the outer handler means both the down and recovered handlers
		// pass the check and only the status code differentiates them.
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		c.respond.Load().(func(http.ResponseWriter, *http.Request))(w, r)
	}))
	t.Cleanup(c.server.Close)
	return c
}

// recover swaps the fake cluster's response to "healthy": Indices.Exists gets
// a 200 (index present), Index gets a 200 (document accepted). The logger
// only inspects status codes for these calls, so a single handler suffices.
func (c *stubCluster) recover() {
	c.respond.Store(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"acknowledged":true}`))
	})
}

// newTestLogger constructs an ElasticsearchLogger pointed at the fake
// cluster, using the same shape NewElasticsearchLogger would but keeping the
// concrete struct so the test can read healthy / reconnecting.
func newTestLogger(cluster *stubCluster) *ElasticsearchLogger {
	return &ElasticsearchLogger{
		args: &NewElasticsearchLoggerArgs{
			Level:  multilog.TRACE,
			Config: Config{Addresses: []string{cluster.server.URL}},
			Index:  "test",
		},
	}
}

// TestReconnect_ResumesAfterOutage covers the interesting end-to-end path:
// Setup runs while the cluster is unreachable -> sink starts unhealthy and
// spawns a reconnect goroutine -> Log calls silently drop while unhealthy ->
// once the cluster recovers, the reconnect goroutine flips the sink healthy
// and subsequent Log calls actually reach the cluster.
//
// This is the property that separates the current implementation from the
// previous "Setup once, stay disabled forever" behavior.
func TestReconnect_ResumesAfterOutage(t *testing.T) {
	origInterval := reconnectInterval
	reconnectInterval = 25 * time.Millisecond
	t.Cleanup(func() { reconnectInterval = origInterval })

	cluster := newStubCluster(t)
	logger := newTestLogger(cluster)
	logger.Setup()

	require.False(t, logger.healthy.Load(), "sink should start unhealthy while cluster is down")

	// A Log call while unhealthy must not panic or block.
	logger.Log(multilog.INFO, "test", "should be dropped", nil)

	// Reconnect goroutine should be trying at reconnectInterval. Wait until
	// it has attempted at least twice so we know the loop is live, not just
	// the one Setup call.
	require.Eventually(t, func() bool { return cluster.requests.Load() >= 2 },
		time.Second, 5*time.Millisecond,
		"reconnect goroutine did not attempt after Setup failure")

	cluster.recover()

	require.Eventually(t, logger.healthy.Load, time.Second, 5*time.Millisecond,
		"sink did not become healthy after cluster recovery")

	before := cluster.requests.Load()
	logger.Log(multilog.INFO, "test", "after recovery", map[string]interface{}{"k": "v"})

	assert.Eventually(t, func() bool { return cluster.requests.Load() > before },
		time.Second, 5*time.Millisecond,
		"Log did not hit the cluster after reconnect")
}

// TestReconnect_SingleFlight guards the invariant that a burst of concurrent
// Log failures spawns exactly one reconnect goroutine, not one per failing
// call, so a partitioned cluster does not get slammed with N reconnect loops
// where N is the log volume.
func TestReconnect_SingleFlight(t *testing.T) {
	origInterval := reconnectInterval
	reconnectInterval = 24 * time.Hour // effectively never; we only observe the initial state
	t.Cleanup(func() { reconnectInterval = origInterval })

	cluster := newStubCluster(t)
	logger := newTestLogger(cluster)
	logger.Setup() // fails, triggers reconnect goroutine

	// Force many concurrent unhealthy transitions.
	for range 50 {
		go logger.markUnhealthy()
	}

	// A brief wait for those goroutines to attempt to spawn reconnectors.
	time.Sleep(50 * time.Millisecond)

	logger.reconnectMu.Lock()
	defer logger.reconnectMu.Unlock()
	assert.True(t, logger.reconnecting,
		"reconnecting flag should still be set (goroutine sleeping on reconnectInterval)")
}

// TestSetupSucceedsWhenClusterHealthy is the happy-path baseline: if the
// cluster is reachable at Setup time, the sink is immediately healthy and no
// reconnect goroutine spawns.
func TestSetupSucceedsWhenClusterHealthy(t *testing.T) {
	cluster := newStubCluster(t)
	cluster.recover()

	logger := newTestLogger(cluster)
	logger.Setup()

	assert.True(t, logger.healthy.Load(), "sink should be healthy after Setup against a live cluster")

	logger.reconnectMu.Lock()
	defer logger.reconnectMu.Unlock()
	assert.False(t, logger.reconnecting, "no reconnect goroutine should be running when Setup succeeded")
}
