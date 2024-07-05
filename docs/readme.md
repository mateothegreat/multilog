# logging with golang

This package provides the ability to use multiple logging output methods.

![alt text](<CleanShot 2024-07-04 at 19.28.48.png>)

## Defining a custom logger

```go
package main

import (
	"log"

	"github.com/mateothegreat/go-multilog/logging"
	"github.com/mateothegreat/go-multilog/logging/loggers"
	"github.com/mateothegreat/go-multilog/logging/types"
)

type CustomLogData struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

func init() {
	types.RegisterLogger(types.LogMethod("console"), loggers.NewConsoleLogger(&loggers.NewConsoleLoggerArgs{
		Format: loggers.FormatText,
	}))

	types.RegisterLogger(types.LogMethod("elasticsearch"), loggers.NewElasticsearchLogger(&loggers.NewElasticsearchLoggerArgs{
		Addresses:          []string{"https://localhost:9200"},
		Username:           "elastic",
		Password:           "elastic",
		Index:              "logs-1",
		InsecureSkipVerify: true,
	}))

	types.RegisterLogger(types.LogMethod("customerLogger1"), &types.CustomLogger{
		Log: func(level types.LogLevel, group string, message string, v any) {
			log.Printf("logged via customerLogger1: %s: %s", group, message)
		},
	})

	// Register a custom logger:
	customLogger1 := types.NewLogger(types.LogMethod("customerLogger2"))
	// If needed, you can do stuff here when the logger is setup such as
	// connecting to something like elasticsearch or whatever:
	customLogger1.Setup = func() {
		log.Println("Setup customerLogger2")
	}
	// Define the log method:
	customLogger1.Log = func(level types.LogLevel, group string, message string, v any) {
		log.Printf("logged via customerLogger: %s: %s", group, message)
	}
}

func main() {
	logging.Debug("my_package_name", "test", CustomLogData{
		Foo: "foo",
		Bar: 1,
	})
	logging.Warn("my_package_name", "it's about to explode...", CustomLogData{
		Foo: "boom",
		Bar: 1234234234234,
	})

	logging.Error("my_package_name", "some error!", CustomLogData{
		Foo: "bad things happened bro",
		Bar: 123,
	})
}
```

![alt text](<CleanShot 2024-07-04 at 19.03.19.png>)