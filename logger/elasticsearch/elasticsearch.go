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

// Setup initializes the Elasticsearch client and best-effort creates the
// target index.
//
// Errors are logged and cause the logger to degrade to a no-op sink rather
// than crashing the process. A telemetry sink must never take the application
// down: an unreachable Elasticsearch during pod startup or a bad index mapping
// is a monitoring problem, not a service problem, and killing the process
// removes the operator's only remaining channel (stderr and other registered
// loggers) for observing the outage.
//
// When the client cannot be constructed or the index cannot be verified, l.client
// stays nil and Log becomes a no-op. Callers that need hard-fail behavior on
// bad configuration should validate the config before calling Setup.
func (l *ElasticsearchLogger) Setup() {
	client, err := elasticsearch.NewClient(l.args.Config)
	if err != nil {
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

	// Best-effort index bootstrap. A failure here (unreachable ES, missing
	// permissions, DNS blip during startup) disables the sink but keeps the
	// process running so other loggers still receive events.
	existsRes, err := l.client.Indices.Exists([]string{l.args.Index})
	if err != nil {
		log.Printf("elasticsearch logger: could not check if index %q exists, disabling sink: %s", l.args.Index, err)
		l.client = nil
		return
	}
	defer existsRes.Body.Close()

	if existsRes.StatusCode != 404 {
		return
	}
	if l.args.Mapping != "" {
		createRes, err := l.client.Indices.Create(l.args.Index,
			l.client.Indices.Create.WithBody(bytes.NewReader([]byte(l.args.Mapping))))
		if err != nil {
			log.Printf("elasticsearch logger: could not create index %q with mapping, disabling sink: %s", l.args.Index, err)
			l.client = nil
			return
		}
		defer createRes.Body.Close()
		if createRes.IsError() {
			log.Printf("elasticsearch logger: create index %q returned error, disabling sink: %s", l.args.Index, createRes.String())
			l.client = nil
			return
		}
		return
	}
	createRes, err := l.client.Indices.Create(l.args.Index)
	if err != nil {
		log.Printf("elasticsearch logger: could not create index %q, disabling sink: %s", l.args.Index, err)
		l.client = nil
		return
	}
	defer createRes.Body.Close()
	if createRes.IsError() {
		log.Printf("elasticsearch logger: create index %q returned error, disabling sink: %s", l.args.Index, createRes.String())
		l.client = nil
	}
}

// Log ships one message to Elasticsearch.
//
// Transport failures are logged to stderr and dropped: a logging sink must
// never crash the process it observes. Prior versions called log.Fatalf on
// any indexing error, which turned a network partition into an application
// outage -- an app that logs "database unreachable" during a blip would kill
// itself trying to record that fact.
//
// A nil client (Setup failed or the sink was disabled) drops the message
// silently. The console logger, if registered alongside, still receives it.
func (l *ElasticsearchLogger) Log(level multilog.LogLevel, group string, message string, v map[string]interface{}) {
	if l.client == nil {
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
		log.Printf("elasticsearch logger: could not marshal document, dropping: %s", err)
		return
	}

	res, err := l.client.Index(l.args.Index, bytes.NewReader(data))
	if err != nil {
		log.Printf("elasticsearch logger: could not index document, dropping: %s", err)
		return
	}
	defer res.Body.Close()
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
