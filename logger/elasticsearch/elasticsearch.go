package elasticsearch

import (
	"bytes"
	"encoding/json"
	"log"
	"regexp"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mateothegreat/multilog"
)

// DefaultMapping is the default mapping for the elasticsearch index.
const DefaultMapping = `
	{
		"mappings": {
			"properties": {
				"time": { "type": "date" },
				"level": { "type": "keyword" },
				"group": { "type": "keyword" },
				"message": { "type": "text" },
				"data": { "type": "object" }
			}
		}
	}`

// reconnectInterval is how long the background reconnect goroutine waits
// between attempts to re-verify the Elasticsearch index after Log has marked
// the sink unhealthy.
//
// A fixed interval, rather than exponential backoff, matches the failure
// modes this exists for: managed-Elasticsearch failovers and short DNS/network
// blips typically resolve within tens of seconds, and only one reconnect
// goroutine runs at a time, so a longer backoff would delay recovery without
// reducing load on the cluster.
//
// It is a var so tests can shorten it rather than waiting out the real
// interval.
var reconnectInterval = 3 * time.Second

// Setup initializes the Elasticsearch client, compiles filter patterns, and
// best-effort ensures the target index exists.
//
// Errors are logged and put the sink into an unhealthy state that a
// background goroutine keeps retrying. A telemetry sink must never take the
// application down: an unreachable Elasticsearch during pod startup or a bad
// index mapping is a monitoring problem, not a service problem, and killing
// the process removes the operator's only remaining channel for observing
// the outage.
func (l *ElasticsearchLogger) Setup() {
	client, err := elasticsearch.NewClient(l.args.Config)
	if err != nil {
		// Client construction failures are configuration errors (bad TLS
		// setup, invalid addresses) that will not resolve on their own, so
		// there is no reconnect loop to spawn -- the sink stays disabled.
		log.Printf("elasticsearch logger: could not create client, disabling sink: %s", err)
		return
	}
	l.client = client

	// Compile the filter patterns if provided. A bad pattern is skipped
	// rather than fatal: dropping a filter is safer than dropping the
	// process, and the offending pattern is logged so the operator can fix
	// it.
	for _, pattern := range l.args.FilterDropPatterns {
		if pattern == nil {
			continue
		}
		compiledPattern, err := regexp.Compile(*pattern)
		if err != nil {
			log.Printf("elasticsearch logger: skipping invalid filter pattern %q: %s", *pattern, err)
			continue
		}
		l.filterPatterns = append(l.filterPatterns, compiledPattern)
	}

	if err := l.bootstrapIndex(); err != nil {
		log.Printf("elasticsearch logger: index bootstrap failed, will retry in background: %s", err)
		l.triggerReconnect()
		return
	}
	l.healthy.Store(true)
}

// Log ships one message to Elasticsearch.
//
// Transport failures mark the sink unhealthy, spawn a background reconnect
// goroutine (if one is not already running), and drop the message. Subsequent
// calls skip the network round-trip until reconnect succeeds, so a partitioned
// cluster does not flood stderr with per-message errors. Other registered
// loggers (console, etc.) still receive the event.
//
// A logging sink must never crash the process it observes: prior versions
// called log.Fatalf on any indexing error, which turned a network partition
// into an application outage.
func (l *ElasticsearchLogger) Log(level multilog.LogLevel, group string, message string, v map[string]interface{}) {
	if l.client == nil || !l.healthy.Load() {
		return
	}
	// Check if the log level is sufficient to log the message.
	if level < l.args.Level {
		return // Drop the message if the log level is lower than the configured level.
	}

	// Check if the message matches any of the filter patterns.
	for _, pattern := range l.filterPatterns {
		if pattern.MatchString(group) || pattern.MatchString(message) {
			return // Drop the message if it matches any of the filter patterns.
		}
	}

	data, err := json.Marshal(ElasticsearchLog{
		Time:    time.Now(),
		Level:   level,
		Group:   group,
		Message: message,
		Data:    v,
	})
	if err != nil {
		// Marshalling failures are payload bugs, not cluster problems, so
		// they do not mark the sink unhealthy: shipping the next message
		// might still work.
		log.Printf("elasticsearch logger: could not marshal document, dropping: %s", err)
		return
	}

	res, err := l.client.Index(l.args.Index, bytes.NewReader(data))
	if err != nil {
		log.Printf("elasticsearch logger: could not index document, marking sink unhealthy: %s", err)
		l.markUnhealthy()
		return
	}
	defer res.Body.Close()
	if res.IsError() {
		log.Printf("elasticsearch logger: index request returned %s, marking sink unhealthy", res.Status())
		l.markUnhealthy()
	}
}

// markUnhealthy transitions the sink to the disabled state and, if no
// reconnect goroutine is already running, starts one.
//
// It is safe to call from many goroutines at once: only the first caller to
// see reconnecting == false spawns the goroutine, so a burst of concurrent
// Log failures produces exactly one background retryer.
func (l *ElasticsearchLogger) markUnhealthy() {
	l.healthy.Store(false)
	l.triggerReconnect()
}

// triggerReconnect starts the reconnect goroutine unless one is already
// running. Callers do not need to hold reconnectMu themselves.
func (l *ElasticsearchLogger) triggerReconnect() {
	l.reconnectMu.Lock()
	defer l.reconnectMu.Unlock()
	if l.reconnecting {
		return
	}
	l.reconnecting = true
	go l.reconnect()
}

// reconnect re-attempts the index bootstrap on reconnectInterval until it
// succeeds, then flips the sink back to healthy. Exactly one instance runs
// at a time, guarded by reconnectMu and the reconnecting flag.
//
// There is no maximum retry count on purpose: a monitoring sink that stops
// trying to reach its cluster after N minutes silently degrades to no
// observability for the rest of the process lifetime, which is worse than
// keeping a low-cost heartbeat running until the network comes back.
func (l *ElasticsearchLogger) reconnect() {
	for {
		time.Sleep(reconnectInterval)
		if err := l.bootstrapIndex(); err != nil {
			log.Printf("elasticsearch logger: reconnect attempt failed, will retry: %s", err)
			continue
		}
		l.healthy.Store(true)
		l.reconnectMu.Lock()
		l.reconnecting = false
		l.reconnectMu.Unlock()
		log.Printf("elasticsearch logger: reconnected to cluster, resuming sink")
		return
	}
}

// bootstrapIndex verifies the configured index exists and creates it (with
// the configured mapping when set) if it does not.
//
// It is called from Setup and from the reconnect goroutine, so it must be
// safe to invoke repeatedly against an already-bootstrapped cluster: the
// exists check is idempotent, and the create branch only runs on a 404.
func (l *ElasticsearchLogger) bootstrapIndex() error {
	existsRes, err := l.client.Indices.Exists([]string{l.args.Index})
	if err != nil {
		return err
	}
	defer existsRes.Body.Close()

	// 200 = index exists; 404 = we need to create it. Anything else
	// (401/403 auth failures, 5xx cluster errors, redirect surprises) means
	// the cluster is not in a state where we can trust its answer, so bail
	// and let the reconnect goroutine try again after a delay.
	if existsRes.StatusCode == 200 {
		return nil
	}
	if existsRes.StatusCode != 404 {
		return &createError{status: existsRes.String()}
	}
	if l.args.Mapping != "" {
		createRes, err := l.client.Indices.Create(l.args.Index,
			l.client.Indices.Create.WithBody(bytes.NewReader([]byte(l.args.Mapping))))
		if err != nil {
			return err
		}
		defer createRes.Body.Close()
		if createRes.IsError() {
			return &createError{status: createRes.String()}
		}
		return nil
	}
	createRes, err := l.client.Indices.Create(l.args.Index)
	if err != nil {
		return err
	}
	defer createRes.Body.Close()
	if createRes.IsError() {
		return &createError{status: createRes.String()}
	}
	return nil
}

// createError wraps a non-2xx create-index response so bootstrapIndex can
// signal it the same way as a transport error, keeping the retry loop in
// reconnect straightforward.
type createError struct {
	status string
}

func (e *createError) Error() string {
	return "elasticsearch: create index returned " + e.status
}

// NewElasticsearchLogger creates a new elasticsearch logger.
//
// Arguments:
//   - args <*NewElasticsearchLoggerArgs>: The arguments to create a new elasticsearch logger.
//
// Returns:
//   - *CustomLogger: The custom logger.
func NewElasticsearchLogger(args *NewElasticsearchLoggerArgs) *multilog.CustomLogger {
	logger := &ElasticsearchLogger{
		args: args,
	}

	return &multilog.CustomLogger{
		Setup: logger.Setup,
		Log:   logger.Log,
	}
}
