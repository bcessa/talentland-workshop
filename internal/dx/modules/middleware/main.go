package middleware

import (
	"net/http"

	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/errors"
	mwCors "go.bryk.io/pkg/net/middleware/cors"
	mwGzip "go.bryk.io/pkg/net/middleware/gzip"
	mwHeaders "go.bryk.io/pkg/net/middleware/headers"
	mwHSTS "go.bryk.io/pkg/net/middleware/hsts"
	mwMetadata "go.bryk.io/pkg/net/middleware/metadata"
	mwProxy "go.bryk.io/pkg/net/middleware/proxy"
	mwRate "go.bryk.io/pkg/net/middleware/rate"
	mwRecovery "go.bryk.io/pkg/net/middleware/recovery"
	mwOtel "go.bryk.io/pkg/otel/http"
)

// Module to manage common HTTP middleware functions.
type Module struct {
	Proxy    bool                `json:"proxy_protocol" yaml:"proxy_protocol" mapstructure:"proxy_protocol"`
	Recovery bool                `json:"panic_recovery" yaml:"panic_recovery" mapstructure:"panic_recovery"`
	Otel     *otelSettings       `json:"otel" yaml:"otel" mapstructure:"otel"`
	Gzip     int                 `json:"gzip" yaml:"gzip" mapstructure:"gzip"`
	Cors     *mwCors.Options     `json:"cors" yaml:"cors" mapstructure:"cors"`
	Headers  map[string]string   `json:"headers" yaml:"headers" mapstructure:"headers"`
	Metadata *mwMetadata.Options `json:"metadata" yaml:"metadata" mapstructure:"metadata"`
	Hsts     *mwHSTS.Options     `json:"hsts" yaml:"hsts" mapstructure:"hsts"`
	Rate     *rateSettings       `json:"rate" yaml:"rate" mapstructure:"rate"`
}

// Handler defines the common signature for middleware functions.
type Handler = func(http.Handler) http.Handler

// Name returns the default module identifier: "middleware".
func (m *Module) Name() string {
	return "middleware"
}

// Load configuration settings from the provided viper instance.
func (m *Module) Load(v *viper.Viper) error {
	return v.Unmarshal(&m)
}

// Flags returns no CLI options by default.
func (m *Module) Flags(_ string) []cli.Param {
	return []cli.Param{}
}

// Customize the provided `*[]func(http.Handler) http.Handler` target.
func (m *Module) Customize(target any) error {
	// ensure provide target is of correct type
	opts, ok := target.(*[]Handler)
	if !ok {
		return errors.New("target must be of type `*[]func(http.Handler) http.Handler`")
	}

	nOpts := []Handler{}
	if m.Proxy {
		nOpts = append(nOpts, mwProxy.Handler())
	}
	if m.Gzip > 0 {
		nOpts = append(nOpts, mwGzip.Handler(m.Gzip))
	}
	if len(m.Headers) > 0 {
		nOpts = append(nOpts, mwHeaders.Handler(m.Headers))
	}
	if m.Metadata != nil {
		nOpts = append(nOpts, mwMetadata.Handler(*m.Metadata))
	}
	if m.Cors != nil {
		nOpts = append(nOpts, mwCors.Handler(*m.Cors))
	}
	if m.Otel != nil && m.Otel.Enabled {
		mwOtelOpts := []mwOtel.Option{}
		if m.Otel.NetworkEvents {
			mwOtelOpts = append(mwOtelOpts, mwOtel.WithNetworkEvents())
		}
		if m.Otel.TraceHeader != "" {
			mwOtelOpts = append(mwOtelOpts, mwOtel.WithTraceInHeader(m.Otel.TraceHeader))
		}
		if len(m.Otel.OmitPaths) > 0 {
			mwOtelOpts = append(mwOtelOpts, mwOtel.WithFilter(mwOtel.FilterByPath(m.Otel.OmitPaths)))
		}
		nOpts = append(nOpts, mwOtel.NewMonitor(mwOtelOpts...).ServerMiddleware())
	}
	if m.Hsts != nil {
		nOpts = append(nOpts, mwHSTS.Handler(*m.Hsts))
	}
	if m.Rate != nil {
		nOpts = append(nOpts, mwRate.Handler(m.Rate.Limit, m.Rate.Burst))
	}
	if m.Recovery {
		nOpts = append(nOpts, mwRecovery.Handler())
	}

	// adjust target
	*opts = append(*opts, nOpts...)
	return nil
}

type rateSettings struct {
	Limit uint `json:"limit" yaml:"limit" mapstructure:"limit"`
	Burst uint `json:"burst" yaml:"burst" mapstructure:"burst"`
}

type otelSettings struct {
	Enabled       bool     `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	NetworkEvents bool     `json:"network_events" yaml:"network_events" mapstructure:"network_events"`
	TraceHeader   string   `json:"trace_header" yaml:"trace_header" mapstructure:"trace_header"`
	OmitPaths     []string `json:"omit_paths" yaml:"omit_paths" mapstructure:"omit_paths"`
}
