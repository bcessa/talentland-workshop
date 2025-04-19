/*
Package otel provides a `dx` module to manage an `otel.Operator` instance.

This module expects a configuration source like:

	otel:
		service_name: "sample_service"
		service_version: "0.1.0"
		metrics_host: true
		metrics_runtime: true
		collector:
			endpoint: "" # if empty, output will be discarded
			protocol: "grpc" # grpc or http
		attributes:
			environment: dev
		sentry:
			dsn: "" # if empty, output will be discarded
			environment: dev
*/
package otel
