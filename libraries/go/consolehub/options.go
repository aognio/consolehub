package consolehub

import (
	"os"

	"github.com/aognio/consolehub/libraries/go/consolehub/config"
	"github.com/aognio/consolehub/libraries/go/consolehub/transport"
	"github.com/aognio/consolehub/libraries/go/consolehub/ulid"
)

// Options holds configuration settings for the ConsoleHub client.
type Options struct {
	Endpoint      string
	Token         string
	Tenant        string
	App           string
	ClientRunID   string // Mandatory ULID generated at the client
	Hostname      string // Mandatory: System hostname (auto-detected if not specified)
	PID           int    // Mandatory: Process ID (auto-detected if not specified)
	AppVersion    string // Optional: Application version string (e.g. "1.2.0")
	OSName        string // Optional: Operating System name (e.g. "linux", "darwin")
	QueueCapacity int
	Disabled      bool
	Transport     transport.Transport
	Env           config.Environment
}

type Option func(*Options)

func defaultOptions() Options {
	env := config.AutoDetect()

	endpoint := config.GetEnvOrDefault("CONSOLEHUB_ENDPOINT", "ws://localhost:3787/api/v1/rpc/ws")
	token := config.GetEnvOrDefault("CONSOLEHUB_TOKEN", "")
	tenant := config.GetEnvOrDefault("CONSOLEHUB_TENANT", "default")
	app := config.GetEnvOrDefault("CONSOLEHUB_APP", env.CommandLine)
	if app == "" {
		app = "go-app"
	}
	disEnv := os.Getenv("CONSOLEHUB_DISABLED")
	disabled := disEnv == "true" || disEnv == "1" || disEnv == "yes"

	return Options{
		Endpoint:      endpoint,
		Token:         token,
		Tenant:        tenant,
		App:           app,
		ClientRunID:   ulid.Make(), // Mandatory client-side ULID generation
		Hostname:      env.Hostname, // Mandatory auto-detection fallback
		PID:           env.PID,      // Mandatory auto-detection fallback
		AppVersion:    "",           // Optional
		OSName:        env.Platform, // Optional OS Name
		QueueCapacity: 10000,
		Disabled:      disabled,
		Env:           env,
	}
}

// WithEndpoint sets the JSON-RPC WebSocket server endpoint.
func WithEndpoint(endpoint string) Option {
	return func(o *Options) {
		if endpoint != "" {
			o.Endpoint = endpoint
		}
	}
}

// WithToken sets the client authentication bearer token.
func WithToken(token string) Option {
	return func(o *Options) {
		o.Token = token
	}
}

// WithTenant sets the target tenant slug/ID.
func WithTenant(tenant string) Option {
	return func(o *Options) {
		if tenant != "" {
			o.Tenant = tenant
		}
	}
}

// WithApp sets the application identifier name.
func WithApp(app string) Option {
	return func(o *Options) {
		if app != "" {
			o.App = app
		}
	}
}

// WithClientRunID sets or overrides the mandatory client-side ULID run identifier.
func WithClientRunID(clientRunID string) Option {
	return func(o *Options) {
		if clientRunID != "" {
			o.ClientRunID = clientRunID
		}
	}
}

// WithHostname sets the mandatory host machine hostname.
func WithHostname(hostname string) Option {
	return func(o *Options) {
		if hostname != "" {
			o.Hostname = hostname
		}
	}
}

// WithPID sets the mandatory process ID.
func WithPID(pid int) Option {
	return func(o *Options) {
		if pid > 0 {
			o.PID = pid
		}
	}
}

// WithAppVersion sets the optional application version string.
func WithAppVersion(version string) Option {
	return func(o *Options) {
		o.AppVersion = version
	}
}

// WithOSName sets the optional operating system name (e.g. "linux", "darwin").
func WithOSName(osName string) Option {
	return func(o *Options) {
		o.OSName = osName
	}
}

// WithQueueCapacity sets the maximum bounded in-memory queue depth.
func WithQueueCapacity(cap int) Option {
	return func(o *Options) {
		if cap > 0 {
			o.QueueCapacity = cap
		}
	}
}

// WithDisabled disables network transmission for local-only execution.
func WithDisabled(disabled bool) Option {
	return func(o *Options) {
		o.Disabled = disabled
	}
}

// WithTransport injects a custom or mock transport implementation.
func WithTransport(trans transport.Transport) Option {
	return func(o *Options) {
		o.Transport = trans
	}
}

// GetClientRunID returns the mandatory client-side ULID run identifier.
func (o *Options) GetClientRunID() string {
	if o.ClientRunID == "" {
		o.ClientRunID = ulid.Make()
	}
	return o.ClientRunID
}
