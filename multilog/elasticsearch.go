package multilog

import (
	"bytes"
	"encoding/json"
	"log"
	"regexp"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

// ElasticsearchLog is the structure of the log that will be sent to the elasticsearch cluster.
// Data can be any serializable type.
type ElasticsearchLog struct {
	Level   LogLevel  `json:"level"`
	Group   string    `json:"group"`
	Message string    `json:"message"`
	Data    any       `json:"data"`
	Time    time.Time `json:"time"`
}

// NewElasticsearchLoggerArgs are the arguments to create a new elasticsearch logger.
type NewElasticsearchLoggerArgs struct {
	// Config is the configuration for the elasticsearch client. https://www.elastic.co/guide/en/elasticsearch/client/go-api/current/connecting.html
	Config elasticsearch.Config
	// Index is the index to use to send the logs to.
	Index string
	// Mapping is the mapping for the index.
	Mapping *string
	// FilterDropPatterns is a slice of regex patterns to filter out log messages.
	FilterDropPatterns []*string
}

// ElasticsearchLogger is the logger that sends logs to an elasticsearch cluster.
type ElasticsearchLogger struct {
	args           *NewElasticsearchLoggerArgs
	client         *elasticsearch.Client
	filterPatterns []*regexp.Regexp
}

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

	// If the mapping is not provided, we assume that the index already exists.
	if l.args.Mapping == nil {
		// Check if the index already exists.
		existsRes, err := l.client.Indices.Exists([]string{l.args.Index})
		if err != nil {
			log.Fatalf("error checking if index exists: %s", err)
		}
		defer existsRes.Body.Close()

		// Index does not exist, create it.
		if existsRes.StatusCode == 404 {
			if l.args.Mapping != nil {
				createRes, err := l.client.Indices.Create(l.args.Index, l.client.Indices.Create.WithBody(bytes.NewReader([]byte(*l.args.Mapping))))
				if err != nil {
					log.Fatalf("error creating index with mapping: %s", err)
				}
				defer createRes.Body.Close()

				if createRes.IsError() {
					log.Fatalf("error response from creating index: %s", createRes.String())
				}
			}
		}
	}

}

// Log is the method to log a message to the elasticsearch cluster.
func (l *ElasticsearchLogger) Log(level LogLevel, group string, message string, v any) {
	// Check if the message matches any of the filter patterns.
	for _, pattern := range l.filterPatterns {
		if pattern.MatchString(message) {
			log.Printf("dropping message due to filter pattern: %s", message)
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
func NewElasticsearchLogger(args *NewElasticsearchLoggerArgs) *CustomLogger {
	logger := &ElasticsearchLogger{
		args: args,
	}

	return &CustomLogger{
		Setup: logger.Setup,
		Log:   logger.Log,
	}
}
