/*
Package tls provides a `dx` module to manage common TLS settings.

This module expects a configuration source like:

	tls:
		enabled: false
		system_ca: true
		cert: testdata/server.sample_cer
		key: testdata/server.sample_key
		custom_ca: []
		auth_ca: []
*/
package tls
