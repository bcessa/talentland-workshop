package tls

import (
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/errors"
)

// Module to manage common TLS settings.
type Module struct {
	Enabled  bool     `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	SystemCA bool     `json:"system_ca" yaml:"system_ca" mapstructure:"system_ca"`
	Cert     string   `json:"cert" yaml:"cert" mapstructure:"cert"`
	Key      string   `json:"key" yaml:"key" mapstructure:"key"`
	CustomCA []string `json:"custom_ca" yaml:"custom_ca" mapstructure:"custom_ca"`
	AuthCA   []string `json:"auth_ca" yaml:"auth_ca" mapstructure:"auth_ca"`

	// private expanded values
	cert      []byte
	key       []byte
	customCAs [][]byte
	authCAs   [][]byte
}

// Settings loaded/managed by the module.
type Settings struct {
	// Whether to use CAs enabled in the local system.
	SystemCAs bool

	// x509 certificate.
	Certificate []byte

	// Private key corresponding to the x509 certificate.
	PrivateKey []byte

	// Custom CAs used, if any.
	CustomCAs [][]byte

	// Custom CAs used for client authentication, if any.
	AuthCAs [][]byte
}

// Name returns the default module identifier: "tls".
func (m *Module) Name() string {
	return "tls"
}

// Load configuration settings from the provided viper instance.
func (m *Module) Load(v *viper.Viper) error {
	return v.Unmarshal(&m)
}

// Flags exposes core server settings as CLI flags.
func (m *Module) Flags(_ string) []cli.Param {
	return []cli.Param{}
}

// Customize is not supported by the module.
func (m *Module) Customize(_ any) error {
	return errors.New("invalid operation on 'tls' module")
}

// Provide the TLS settings loaded by the module.
func (m *Module) Provide() (*Settings, error) {
	if err := expandTLS(m); err != nil {
		return nil, err
	}
	return &Settings{
		SystemCAs:   m.SystemCA,
		Certificate: m.cert,
		PrivateKey:  m.key,
		CustomCAs:   m.customCAs,
	}, nil
}
