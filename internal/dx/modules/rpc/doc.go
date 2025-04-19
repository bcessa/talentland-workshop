/*
Package rpc provides a `dx` module to manage an `rpc.Server` instance.

This module expects a configuration source like:

	rpc:
		port: 9090
		network_interface: all
		unix_socket: ""
		input_validation: true
		reflection: true
		resource_limits:
			connections: 1000
			requests: 50
			rate: 500
		tls: {}
		http:
			enabled: true
			middleware: {}
*/
package rpc
