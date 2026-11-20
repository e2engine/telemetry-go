package telemetry_go

import (
	"context"
	"os"
	"runtime"
	"time"
	"uuid"

	"github.com/ygrebnov/errorc"

	"github.com/e2engine/telemetry-go/internal/core"
	"github.com/e2engine/telemetry-go/pkg/errors"
	"github.com/e2engine/telemetry-go/pkg/keys"
)

// defaultBaseURL is the default base URL for the telemetry service.
var defaultBaseURL string // injected at build time via ldflags

// Client sends telemetry synchronously and is safe for concurrent use.
type Client struct {
	disabled bool
	service  *core.Service
}

// Config holds the configuration for the telemetry client.
type Config struct {
	Disabled bool `yaml:"disabled,omitempty"`
}

func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.ErrNilConfig
	}

	disabled := resolveDisabled(cfg.Disabled, defaultBaseURL)

	c := &Client{
		disabled: disabled,
	}

	if disabled {
		return c, nil
	}

	c.service = core.NewService(defaultBaseURL)
	return c, nil
}

// IsDisabled reports the enablement state captured when the client was created.
func (c *Client) IsDisabled() bool {
	return c.disabled
}

func (c *Client) Initialize(ctx context.Context) error {
	if c.disabled {
		return nil
	}

	return c.service.Initialize(ctx)
}

const maxBatchSize = 25

// SendEvents sends a batch of events to the telemetry service.
func (c *Client) SendEvents(ctx context.Context, events []Event) error {
	if c.disabled {
		return nil
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if len(events) == 0 || len(events) > maxBatchSize {
		return errorc.With(
			errors.ErrInvalidBatch,
			errorc.Int(keys.EventCount, len(events)),
			errorc.Int(keys.EventLimit, maxBatchSize),
		)
	}

	return c.service.SendEvents(ctx, events)
}

// telemetryDisabledEnv is the telemetry opt-out variable.
const telemetryDisabledEnv = "E2ENGINE_TELEMETRY_DISABLED"

// resolveDisabled reports whether telemetry is disabled.
// Telemetry is opt-out, meaning it is enabled by default.
// Environment variable overrides the configuration file.
func resolveDisabled(disabledInConfig bool, defaultBaseURL string) bool {
	if defaultBaseURL == "" {
		return true
	}

	value, present := os.LookupEnv(telemetryDisabledEnv)

	if !present || value == "" {
		return disabledInConfig
	}

	return value != "0" && value != "false"
}

// NewEvent constructs an event with a fresh event ID and current UTC timestamp.
func NewEvent(
	sessionID uuid.UUID,
	source EventSource,
	operation Operation,
	outcome OperationOutcome,
	duration time.Duration,
	errorCategory OperationErrorCategory,
	errorCode OperationErrorCode,
	version string,
) Event {
	return Event{
		ID:            uuid.New(),
		SchemaVersion: SchemaVersion,
		SessionID:     sessionID,
		Timestamp:     time.Now().UTC(),

		Source:     source,
		Operation:  operation,
		Outcome:    outcome,
		DurationMS: duration.Milliseconds(),

		ErrorCategory: errorCategory,
		ErrorCode:     errorCode,

		OS:      getOS(),
		Arch:    getArch(),
		Version: version,
	}
}

func getOS() OS {
	switch runtime.GOOS {
	case "linux":
		return OSLinux
	case "darwin":
		return OSDarwin
	case "windows":
		return OSWindows
	default:
		return OSUnknown
	}
}

func getArch() Arch {
	switch runtime.GOARCH {
	case "amd64":
		return ArchAMD64
	case "arm64":
		return ArchARM64
	default:
		return ArchUnknown
	}
}
