package multilog

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
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
	// Addresses is the list of addresses to connect to the elasticsearch cluster.
	Addresses []string
	// Username is the username to use to connect to the elasticsearch cluster.
	Username string
	// Password is the password to use to connect to the elasticsearch cluster.
	Password string
	// Index is the index to use to send the logs to.
	Index string
	// InsecureSkipVerify is a flag to skip SSL certificate verification.
	InsecureSkipVerify bool
}

// ElasticsearchLogger is the logger that sends logs to an elasticsearch cluster.
type ElasticsearchLogger struct {
	args   *NewElasticsearchLoggerArgs
	client *elasticsearch.Client
}

// Setup is the method to setup the elasticsearch logger.
func (l *ElasticsearchLogger) Setup() {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: l.args.Addresses,
		Username:  l.args.Username,
		Password:  l.args.Password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: l.args.InsecureSkipVerify,
			},
		},
	})
	if err != nil {
		log.Fatalf("error creating elasticsearch client: %s", err)
	}
	l.client = client
}

// Log is the method to log a message to the elasticsearch cluster.
func (l *ElasticsearchLogger) Log(level LogLevel, group string, message string, v any) {
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
