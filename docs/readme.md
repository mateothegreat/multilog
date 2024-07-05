# logging with golang

This package provides the ability to use multiple logging output methods.

![alt text](<CleanShot 2024-07-04 at 19.28.48.png>)

## Defining a custom logger

```go
package main

import (
	"log"

	"github.com/mateothegreat/go-multilog/logging"
	"github.com/mateothegreat/go-multilog/logging/types"
)

func init() {
	// Register a custom logger:
	customLogger1 := types.NewCustomLogger(types.LogMethod("customerLogger1"))
	// If needed, you can do stuff here when the logger is setup such as
	// connecting to something like elasticsearch or whatever:
	customLogger1.Setup = func() {
		log.Println("Setup customerLogger1")
	}
	// Define the log method:
	customLogger1.Log = func(level types.LogLevel, group string, message string, v any) {
		log.Printf("logged via customerLogger1: %s: %s", group, message)
	}
}

func main() {
	// Log something:
	logging.Debug("test", "test", "test")
}
```

![alt text](<CleanShot 2024-07-04 at 19.03.19.png>)