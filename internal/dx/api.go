package dx

import (
	"sync"

	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	"go.bryk.io/pkg/errors"
)

// Registry instances can be used to manage a collection of
// modules required by an application.
type Registry struct {
	name    string
	modules map[string]Module
	mu      sync.Mutex
}

// NewRegistry returns a new module registry.
func NewRegistry(name string, mods ...Module) *Registry {
	r := &Registry{
		name:    name,
		modules: make(map[string]Module),
	}
	for _, mod := range mods {
		r.Add(mod)
	}
	return r
}

// Load configuration options managed by the provided Viper instance.
func (r *Registry) Load(v *viper.Viper) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for name, mod := range r.modules {
		if err := mod.Load(v); err != nil {
			return errors.Wrapf(err, "failed loading module %s", name)
		}
	}
	return nil
}

// Add (or) replace a module to the registry.
func (r *Registry) Add(mod Module) {
	r.mu.Lock()
	r.modules[mod.Name()] = mod
	r.mu.Unlock()
}

// Get a module from the registry; if no module with the given name
// exists this method returns `nil`.
func (r *Registry) Get(name string) Module {
	r.mu.Lock()
	mod, ok := r.modules[name]
	r.mu.Unlock()
	if !ok {
		return nil
	}
	return mod
}

// Module implementations provide a basic responsibility encapsulation and
// reutilization mechanism for higher-level applications. The most basic
// form of modules simply manage internal state; their interaction with
// external components is limited to customizing a given target.
type Module interface {
	// Unique module identifier.
	Name() string

	// Load module-specific configuration options managed by the provided
	// Viper instance.
	Load(*viper.Viper) error

	// Returns any configuration options that could be exposed directly
	// as CLI flags.
	Flags(appName string) []cli.Param

	// Customize the provided `target` element based on the module's current
	// internal state and specific requirements.
	Customize(target any) error
}

// Provider modules can also initialize specialized components and hand them
// over for consumption.
type Provider interface {
	// "inherit" all the base functions of a simple module
	Module

	// Provide is responsible of returning a reference that was properly
	// initialized by the module implementation. The caller is responsible
	// for any further management tasks related to the resource.
	Provide() (resource any, err error)
}
