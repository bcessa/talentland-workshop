package otel

import (
	"time"

	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/errors"
	otelSdk "go.bryk.io/pkg/otel/sdk"
	"go.bryk.io/pkg/otel/sentry"
)

// Module to manage the settings for an `otel.Operator` instance.
type Module struct {
	conf struct {
		Otel *settings `json:"otel" yaml:"otel" mapstructure:"otel"`
	}
}

// Name returns the default module identifier: "otel".
func (m *Module) Name() string {
	return "otel"
}

// Load configuration settings from the provided viper instance.
func (m *Module) Load(v *viper.Viper) error {
	if m.conf.Otel == nil {
		m.conf.Otel = &settings{Sentry: new(sentry.Options)}
	}
	return v.Unmarshal(&m.conf)
}

// Flags returns no CLI options by default.
func (m *Module) Flags(_ string) []cli.Param {
	return []cli.Param{}
}

// Customize the provided `*[]otel.OperatorOption` target.
func (m *Module) Customize(target any) error {
	// ensure provide target is of correct type
	opts, ok := target.(*[]otelSdk.Option)
	if !ok {
		return errors.New("target must be of type `*[]sdk.Option`")
	}

	// check if module is enabled, return empty options list if not
	if !m.conf.Otel.Enabled {
		return nil
	}

	// expand internal module settings
	nOpts := []otelSdk.Option{
		otelSdk.WithServiceName(m.conf.Otel.ServiceName),
		otelSdk.WithServiceVersion(m.conf.Otel.ServiceVersion),
	}
	if m.conf.Otel.HostMetrics {
		nOpts = append(nOpts, otelSdk.WithHostMetrics())
	}
	if m.conf.Otel.RuntimeMetrics {
		nOpts = append(nOpts, otelSdk.WithRuntimeMetrics(5*time.Second))
	}
	if len(m.conf.Otel.Attributes) > 0 {
		nOpts = append(nOpts, otelSdk.WithResourceAttributes(m.conf.Otel.Attributes))
	}
	if collector := m.conf.Otel.Collector; collector.Endpoint != "" {
		protocol := collector.Protocol
		if protocol == "" {
			protocol = "grpc"
		}
		nOpts = append(nOpts, otelSdk.WithExporterOTLP(collector.Endpoint, true, nil, protocol)...)
	}
	if sentryOpts := m.conf.Otel.Sentry; sentryOpts.DSN != "" {
		rep, err := sentry.NewReporter(sentryOpts)
		if err == nil {
			nOpts = append(nOpts,
				otelSdk.WithSpanProcessor(rep.SpanProcessor()),
				otelSdk.WithPropagator(rep.Propagator()),
			)
		}
	}

	// adjust target
	*opts = append(*opts, nOpts...)
	return nil
}

type settings struct {
	Enabled        bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	ServiceName    string `json:"service_name" yaml:"service_name" mapstructure:"service_name"`
	ServiceVersion string `json:"service_version" yaml:"service_version" mapstructure:"service_version"`
	Collector      struct {
		Endpoint string `json:"endpoint" yaml:"endpoint" mapstructure:"endpoint"`
		Protocol string `json:"protocol" yaml:"protocol" mapstructure:"protocol"`
	} `json:"collector" yaml:"collector" mapstructure:"collector"`
	HostMetrics    bool                   `json:"metrics_host" yaml:"metrics_host" mapstructure:"metrics_host"`
	RuntimeMetrics bool                   `json:"metrics_runtime" yaml:"metrics_runtime" mapstructure:"metrics_runtime"`
	Attributes     map[string]interface{} `json:"attributes" yaml:"attributes" mapstructure:"attributes"`
	Sentry         *sentry.Options        `json:"sentry" yaml:"sentry" mapstructure:"sentry"`
}
