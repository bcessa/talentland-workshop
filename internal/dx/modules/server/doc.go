/*
Package server provides a `dx` module to manage an `http.Server` instance.

This module expects a configuration source like:

	server:
		port: 9090
		idle_timeout: 5
		tls:
			enabled: false
			system_ca: true
			cert: testdata/server.sample_cer
			key: testdata/server.sample_key
			custom_ca: []
		middleware: {}
*/
package server
