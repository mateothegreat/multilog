package elasticsearch

import (
	"regexp"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mateothegreat/multilog"
)

// ElasticsearchLog is the structure of the log that will be sent to the elasticsearch cluster.
// Data can be any serializable type.
type ElasticsearchLog struct {
	Level   multilog.LogLevel `json:"level"`
	Group   string            `json:"group"`
	Message string            `json:"message"`
	Data    any               `json:"data"`
	Time    time.Time         `json:"time"`
}

type Config = elasticsearch.Config

// NewElasticsearchLoggerArgs are the arguments to create a new elasticsearch logger.
type NewElasticsearchLoggerArgs struct {
	// Level is the log level to use.
	Level multilog.LogLevel
	// Config is the configuration for the elasticsearch client. https://www.elastic.co/guide/en/elasticsearch/client/go-api/current/connecting.html
	Config Config
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
