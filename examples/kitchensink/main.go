package main

import (
	"crypto/tls"
	"log"
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
			multilog.PtrString(".*drop.*"), // Drop any message that contains the word "drop"
		},
	}))

	multilog.RegisterLogger(multilog.LogMethod("customerLogger1"), &multilog.CustomLogger{
		Log: func(level multilog.LogLevel, group string, message string, v any) {
			log.Printf("logged via customerLogger1: %s: %s", group, message)
		},
	})

	// Register a custom logger:
	customLogger1 := multilog.NewLogger(multilog.LogMethod("customerLogger2"))
	// If needed, you can do stuff here when the logger is setup such as
	// connecting to something like elasticsearch or whatever:
	customLogger1.Setup = func() {
		log.Println("Setup customerLogger2")
	}
	// Define the log method:
	customLogger1.Log = func(level multilog.LogLevel, group string, message string, v any) {
		log.Printf("logged via customerLogger: %s: %s", group, message)
	}
}

func main() {
	multilog.Debug("my_package_name", "test", CustomLogData{
		Foo: "foo",
		Bar: 1,
	})
	multilog.Warn("my_package_name", "it's about to explode...", CustomLogData{
		Foo: "boom",
		Bar: 1234234234234,
	})

	multilog.Error("my_package_name", "some error!", CustomLogData{
		Foo: "bad things happened bro",
		Bar: 123,
	})

	multilog.Trace("my_package_name", "some verbose info..", CustomLogData{
		Foo: "it's happpeeennning!!!",
		Bar: 234234234,
	})

	multilog.Trace("nobody_cares_about_this", "this message will get dropped by the filters", nil)
	multilog.Error("block_this_group", "this message will get dropped by the filters", nil)
}
