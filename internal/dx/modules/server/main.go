package server

import (
	"fmt"
	"time"

	dxMW "github.com/bcessa/echo-service/internal/dx/modules/middleware"
	dxTLS "github.com/bcessa/echo-service/internal/dx/modules/tls"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/errors"
	xHttp "go.bryk.io/pkg/net/http"
)

const (
	// default TCP port.
	defaultPort int = 9090
)

// Module to manage the settings for an `http.Server` instance.
type Module struct {
	conf struct {
		Server *settings `json:"server" yaml:"server" mapstructure:"server"`
	}
}

// Name returns the default module identifier: "server".
func (m *Module) Name() string {
	return "server"
}

// Load configuration settings from the provided viper instance.
func (m *Module) Load(v *viper.Viper) error {
	if m.conf.Server == nil {
		m.conf.Server = defaultSettings()
	}
	return v.Unmarshal(&m.conf)
}

// Flags exposes core server settings as CLI flags.
func (m *Module) Flags(appName string) []cli.Param {
	return []cli.Param{
		{
			Name:      "port",
			Usage:     "TCP port to use for the server",
			FlagKey:   "server.port",
			ByDefault: defaultPort,
			Short:     "p",
		},
		{
			Name:      "tls",
			Usage:     "enable secure communications using TLS with provided credentials",
			FlagKey:   "server.tls.enabled",
			ByDefault: false,
		},
		{
			Name:      "tls-ca",
			Usage:     "TLS custom certificate authority (path to PEM file)",
			FlagKey:   "server.tls.custom_ca",
			ByDefault: "",
		},
		{
			Name:      "tls-cert",
			Usage:     "TLS certificate (path to PEM file)",
			FlagKey:   "server.tls.cert",
			ByDefault: fmt.Sprintf("/etc/%s/tls/tls.crt", appName),
		},
		{
			Name:      "tls-key",
			Usage:     "TLS private key (path to PEM file)",
			FlagKey:   "server.tls.key",
			ByDefault: fmt.Sprintf("/etc/%s/tls/tls.key", appName),
		},
	}
}

// Customize the provided `*[]go.bryk.io/pkg/net/http.Option` target.
func (m *Module) Customize(target any) error {
	// ensure provide target is of correct type
	opts, ok := target.(*[]xHttp.Option)
	if !ok {
		return errors.New("target must be of type `*[]http.Option`")
	}

	// expand internal module settings
	nOpts := []xHttp.Option{
		xHttp.WithPort(m.conf.Server.Port),
	}
	if idle := m.conf.Server.Idle; idle > 0 {
		nOpts = append(nOpts, xHttp.WithIdleTimeout(time.Duration(idle)*time.Second))
	}
	tlsConf := m.conf.Server.TLS
	if tlsConf != nil && tlsConf.Enabled {
		tc, err := tlsConf.Provide()
		if err != nil {
			return err
		}
		nOpts = append(nOpts, xHttp.WithTLS(xHttp.TLS{
			Cert:             tc.Certificate,
			PrivateKey:       tc.PrivateKey,
			IncludeSystemCAs: tc.SystemCAs,
			CustomCAs:        tc.CustomCAs,
		}))
	}

	// Add server middleware
	if m.conf.Server.Middleware != nil {
		sm := []dxMW.Handler{}
		if err := m.conf.Server.Middleware.Customize(&sm); err != nil {
			return err
		}
		nOpts = append(nOpts, xHttp.WithMiddleware(sm...))
	}

	// adjust target
	*opts = append(*opts, nOpts...)
	return nil
}

// apply minimal default settings.
func defaultSettings() *settings {
	return &settings{Port: defaultPort}
}

// nolint: lll
type settings struct {
	Port       int           `json:"port" yaml:"port" mapstructure:"port"`
	Idle       int           `json:"idle_timeout" yaml:"idle_timeout" mapstructure:"idle_timeout"`
	TLS        *dxTLS.Module `json:"tls" yaml:"tls" mapstructure:"tls"`
	Middleware *dxMW.Module  `json:"middleware" yaml:"middleware" mapstructure:"middleware"`
}
