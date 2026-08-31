module github.com/mateothegreat/multilog

go 1.25.3

require (
	github.com/fatih/color v1.19.0
	github.com/mateothegreat/multilog/logger/elasticsearch v0.0.0-20251023221020-f38f7d591b17
)

replace github.com/mateothegreat/multilog/logger/elasticsearch => ./logger/elasticsearch

require (
	github.com/elastic/elastic-transport-go/v8 v8.7.0 // indirect
	github.com/elastic/go-elasticsearch/v8 v8.19.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	go.opentelemetry.io/otel v1.28.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)
