# Grafana reporter

A simple http service that generates \*.PDF reports from [Grafana](http://grafana.org/) dashboards.

## Requirements

Runtime requirements

- `pdflatex` installed and available in PATH.
- a running Grafana instance that it can connect to

## Development

### Test

The unit tests can be run using the go tool:

    go test -v ./...