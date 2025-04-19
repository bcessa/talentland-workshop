package rpc

import (
	"fmt"

	dxMW "github.com/bcessa/echo-service/internal/dx/modules/middleware"
	dxTLS "github.com/bcessa/echo-service/internal/dx/modules/tls"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/errors"
	mwRecovery "go.bryk.io/pkg/net/middleware/recovery"
	"go.bryk.io/pkg/net/rpc"
)

const (
	// default TCP port.
	defaultPort int = 9090
)

// Module to manage the settings for a `rpc.Server` instance.
type Module struct {
	conf struct {
		RPC *settings `json:"rpc" yaml:"rpc" mapstructure:"rpc"`
	}
}

// Name returns the default module identifier: "rpc".
func (m *Module) Name() string {
	return "rpc"
}

// Load configuration settings from the provided viper instance.
func (m *Module) Load(v *viper.Viper) error {
	if m.conf.RPC == nil {
		m.conf.RPC = defaultSettings()
	}
	return v.Unmarshal(&m.conf)
}

// Flags exposes core server settings as CLI flags.
func (m *Module) Flags(appName string) []cli.Param {
	return []cli.Param{
		{
			Name:      "port",
			Usage:     "TCP port to use for the server",
			FlagKey:   "rpc.port",
			ByDefault: defaultPort,
			Short:     "p",
		},
		{
			Name:      "http",
			Usage:     "enable HTTP access",
			FlagKey:   "rpc.http.enabled",
			ByDefault: false,
			Short:     "H",
		},
		{
			Name:      "tls",
			Usage:     "enable secure communications using TLS with provided credentials",
			FlagKey:   "rpc.tls.enabled",
			ByDefault: false,
		},
		{
			Name:      "tls-ca",
			Usage:     "TLS custom certificate authority (path to PEM file)",
			FlagKey:   "rpc.tls.custom_ca",
			ByDefault: "",
		},
		{
			Name:      "tls-cert",
			Usage:     "TLS certificate (path to PEM file)",
			FlagKey:   "rpc.tls.cert",
			ByDefault: fmt.Sprintf("/etc/%s/tls/tls.crt", appName),
		},
		{
			Name:      "tls-key",
			Usage:     "TLS private key (path to PEM file)",
			FlagKey:   "rpc.tls.key",
			ByDefault: fmt.Sprintf("/etc/%s/tls/tls.key", appName),
		},
	}
}

// Customize the provided `*[]rpc.ServerOption` target.
func (m *Module) Customize(target any) error {
	// ensure provide target is of correct type
	opts, ok := target.(*[]rpc.ServerOption)
	if !ok {
		return errors.New("target must be of type `*[]rpc.ServerOption`")
	}
	// consistency checks
	if m.conf.RPC.Port != 0 && m.conf.RPC.UnixSocket != "" {
		return errors.New("port and unix socket can't be used simultaneously")
	}
	if m.conf.RPC.Port == 0 && m.conf.RPC.UnixSocket == "" {
		return errors.New("either port or unix socket is required")
	}

	// expand internal module settings
	nOpts := []rpc.ServerOption{
		rpc.WithPanicRecovery(),
	}
	if m.conf.RPC.Resources != nil {
		nOpts = append(nOpts, rpc.WithResourceLimits(*m.conf.RPC.Resources))
	}
	if m.conf.RPC.InputValidation {
		nOpts = append(nOpts, rpc.WithInputValidation())
	}
	if m.conf.RPC.Reflection {
		nOpts = append(nOpts, rpc.WithReflection())
	}
	if m.conf.RPC.Port != 0 {
		nOpts = append(nOpts,
			rpc.WithPort(m.conf.RPC.Port),
			rpc.WithNetworkInterface(m.conf.RPC.NetInt),
		)
	} else {
		nOpts = append(nOpts, rpc.WithUnixSocket(m.conf.RPC.UnixSocket))
	}

	var (
		tc  *dxTLS.Settings
		err error
	)
	tlsConf := m.conf.RPC.TLS
	if tlsConf != nil && tlsConf.Enabled {
		tc, err = tlsConf.Provide()
		if err != nil {
			return err
		}
		nOpts = append(nOpts, rpc.WithTLS(rpc.ServerTLSConfig{
			Cert:             tc.Certificate,
			PrivateKey:       tc.PrivateKey,
			IncludeSystemCAs: tc.SystemCAs,
			CustomCAs:        tc.CustomCAs,
		}))
	}

	// setup HTTP gateway
	if m.conf.RPC.HTTP.Enabled {
		gw, err := rpc.NewGateway(m.gatewayOptions(tc)...)
		if err != nil {
			return errors.Wrap(err, "failed to setup HTTP gateway")
		}
		nOpts = append(nOpts, rpc.WithHTTPGateway(gw))
	}

	// adjust target
	*opts = append(*opts, nOpts...)
	return nil
}

func (m *Module) gatewayOptions(tc *dxTLS.Settings) []rpc.GatewayOption {
	// gateway internal client options
	clOpts := []rpc.ClientOption{
		rpc.WithInsecureSkipVerify(), // accept any cert provided
	}
	if tc != nil {
		clOpts = append(clOpts, rpc.WithClientTLS(rpc.ClientTLSConfig{
			IncludeSystemCAs: tc.SystemCAs,
			CustomCAs:        tc.CustomCAs,
		}))
	}
	gwOpts := []rpc.GatewayOption{
		rpc.WithClientOptions(clOpts...),
		rpc.WithHandlerName("http-gateway"),
		rpc.WithPrettyJSON("application/json+pretty"),
	}

	// gateway middleware
	if m.conf.RPC.HTTP.Middleware != nil {
		gm := []dxMW.Handler{}
		if err := m.conf.RPC.HTTP.Middleware.Customize(&gm); err != nil {
			return gwOpts
		}
		for _, mw := range gm {
			gwOpts = append(gwOpts, rpc.WithGatewayMiddleware(mw))
		}
		gwOpts = append(gwOpts, rpc.WithGatewayMiddleware(mwRecovery.Handler()))
	}
	return gwOpts
}

// apply minimal default settings.
func defaultSettings() *settings {
	return &settings{
		Port:   defaultPort,
		NetInt: rpc.NetworkInterfaceLocal,
	}
}

// nolint: lll
type settings struct {
	Port            int                 `json:"port" yaml:"port" mapstructure:"port"`
	NetInt          string              `json:"network_interface" yaml:"network_interface" mapstructure:"network_interface"`
	UnixSocket      string              `json:"unix_socket" yaml:"unix_socket" mapstructure:"unix_socket"`
	InputValidation bool                `json:"input_validation" yaml:"input_validation" mapstructure:"input_validation"`
	Reflection      bool                `json:"reflection" yaml:"reflection" mapstructure:"reflection"`
	Resources       *rpc.ResourceLimits `json:"resource_limits" yaml:"resource_limits" mapstructure:"resource_limits"`
	TLS             *dxTLS.Module       `json:"tls" yaml:"tls" mapstructure:"tls"`
	HTTP            *gwSettings         `json:"http" yaml:"http" mapstructure:"http"`
}

type gwSettings struct {
	Enabled    bool         `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Middleware *dxMW.Module `json:"middleware" yaml:"middleware" mapstructure:"middleware"`
}
