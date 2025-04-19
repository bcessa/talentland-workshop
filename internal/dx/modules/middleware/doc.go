/*
Package middleware provides a `dx` module to manage HTTP middleware.

This module expects a configuration source like:

	middleware:
		# return internal errors if `panic` occurs
		panic_recovery: true
		# support PROXY headers
		proxy_protocol: true
		# between 1 and 9; 0 to disable
		gzip: 0
		# custom headers return on every response
		headers:
			- x-foo: "bar"
		# retain some headers as `context` metadata
		metadata:
			headers:
				- authorization
				- x-api-key
		# OpenTelemetry instrumentation
		otel:
			enabled: true
			network_events: true
			trace_header: "X-Trace-ID"
			omit_paths:
				- /metrics
				- /v1/ping
				- /v1/health
		# rate limiting
		rate:
			limit: 100
			burst: 10
		# settings for: Cross-Origin-Request-Support
		cors:
			max_age: 300
			options_status_code: 200
			allow_credentials: true
			ignore_options: false
			allowed_headers:
				- authorization
				- content-type
				- x-api-key
			allowed_methods:
				- get
				- head
				- post
				- options
			allowed_origins:
				- "*"
			exposed_headers:
				- authorization
				- x-api-key
*/
package middleware
