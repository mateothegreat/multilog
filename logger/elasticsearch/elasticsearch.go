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

// Setup is the method to setup the elasticsearch logger.
func (l *ElasticsearchLogger) Setup() {
	client, err := elasticsearch.NewClient(l.args.Config)
	if err != nil {
		log.Fatalf("error creating elasticsearch client: %s", err)
	}
	l.client = client

	// Compile the filter patterns if provided.
	for _, pattern := range l.args.FilterDropPatterns {
		if pattern != nil {
			compiledPattern, err := regexp.Compile(*pattern)
			if err != nil {
				log.Fatalf("error compiling filter pattern: %s", err)
			}
			l.filterPatterns = append(l.filterPatterns, compiledPattern)
		}
	}

	// Check if the index already exists.
	existsRes, err := l.client.Indices.Exists([]string{l.args.Index})
	if err != nil {
		log.Fatalf("error checking if index exists: %s", err)
	}
	defer existsRes.Body.Close()

	// Index does not exist, create it.
	if existsRes.StatusCode == 404 {
		if l.args.Mapping != "" {
			createRes, err := l.client.Indices.Create(l.args.Index,
				l.client.Indices.Create.WithBody(bytes.NewReader([]byte(l.args.Mapping))))
			if err != nil {
				log.Fatalf("error creating index with mapping: %s", err)
			}
			defer createRes.Body.Close()

			if createRes.IsError() {
				log.Fatalf("error response from creating index: %s", createRes.String())
			}
		} else {
			createRes, err := l.client.Indices.Create(l.args.Index)
			if err != nil {
				log.Fatalf("error creating index: %s", err)
			}
			defer createRes.Body.Close()

			if createRes.IsError() {
				log.Fatalf("error response from creating index: %s", createRes.String())
			}
		}
	}
}

// Log is the method to log a message to the elasticsearch cluster.
func (l *ElasticsearchLogger) Log(level multilog.LogLevel, group string, message string, v map[string]interface{}) {
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
		log.Fatalf("error marshalling document: %s", err)
	}

	res, err := l.client.Index(l.args.Index, bytes.NewReader(data))
	if err != nil {
		log.Fatalf("error indexing document: %s", err)
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
