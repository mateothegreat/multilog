package main

import (
	"crypto/tls"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mateothegreat/go-multilog/multilog"
)

type CustomLogData struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

func init() {
	multilog.RegisterLogger(multilog.LogMethod("console"), multilog.NewConsoleLogger(&multilog.NewConsoleLoggerArgs{
		Format: multilog.FormatText,
		FilterDropPatterns: []*string{
			multilog.PtrString("block_this_group"),
			multilog.PtrString(".*drop.*"), // Drop any message that contains the word "drop"
		},
	}))

	mapping := `
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

	multilog.RegisterLogger(multilog.LogMethod("elasticsearch"), multilog.NewElasticsearchLogger(&multilog.NewElasticsearchLoggerArgs{
		Config: elasticsearch.Config{
			Addresses: []string{"https://localhost:9200"},
			Username:  "elastic",
			Password:  "elastic",
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		Index:   "logs-3",
		Mapping: &mapping,
		FilterDropPatterns: []*string{
			multilog.PtrString("block_this_group"),
			multilog.PtrString(".*drop.*"), // Drop any message that contains the word "drop"
		},
	}))
}

func main() {
	multilog.Trace("nobody_cares_about_this", "this message will get dropped by the filters", nil)
	multilog.Error("block_this_group", "this message will get dropped by the filters", nil)
}
