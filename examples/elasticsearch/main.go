package main

import (
	"github.com/mateothegreat/multilog"
	elasticsearch "github.com/mateothegreat/multilog/logger/elasticsearch"
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

	multilog.RegisterLogger(multilog.LogMethod("elasticsearch"), elasticsearch.NewElasticsearchLogger(&elasticsearch.NewElasticsearchLoggerArgs{
		Level: multilog.TRACE,
		Config: elasticsearch.Config{
			Addresses: []string{"https://localhost:9200"},
		},
		Index:   "logs-3",
		Mapping: &mapping,
	}))
}

func main() {
	multilog.Trace("nobody_cares_about_this", "this message will get dropped by the filters", nil)
	multilog.Error("block_this_group", "this message will get dropped by the filters", nil)
	multilog.Fatal("die", "this will crash", nil)
}
