package telemetry_go

import (
	"context"
	stderrors "errors"
	"os"
	"runtime"
	"testing"
	"time"
	"uuid"

	telemetryerrors "github.com/e2engine/telemetry-go/pkg/errors"
)

func TestNewClientRejectsNilConfig(t *testing.T) {
	client, err := NewClient(nil)
	if client != nil || !stderrors.Is(err, telemetryerrors.ErrNilConfig) {
		t.Fatalf("NewClient(nil) = (%v, %v)", client, err)
	}
}

func TestNewClientResolvesDisabled(t *testing.T) {
	setDefaultBaseURL(t, "https://telemetry.example.test")

	tests := []struct {
		name           string
		configured     bool
		environment    *string
		wantDisabled   bool
		wantServiceNil bool
	}{
		{name: "enabled by default"},
		{name: "disabled in config", configured: true, wantDisabled: true, wantServiceNil: true},
		{name: "empty uses config", configured: true, environment: stringPointer(""), wantDisabled: true, wantServiceNil: true},
		{name: "zero overrides config", configured: true, environment: stringPointer("0")},
		{name: "false overrides config", configured: true, environment: stringPointer("false")},
		{name: "one disables", environment: stringPointer("1"), wantDisabled: true, wantServiceNil: true},
		{name: "true disables", environment: stringPointer("true"), wantDisabled: true, wantServiceNil: true},
		{name: "unknown fails closed", environment: stringPointer("unexpected"), wantDisabled: true, wantServiceNil: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.environment == nil {
				if err := os.Unsetenv(telemetryDisabledEnv); err != nil {
					t.Fatal(err)
				}
			} else {
				t.Setenv(telemetryDisabledEnv, *test.environment)
			}

			client, err := NewClient(&Config{Disabled: test.configured})
			if err != nil {
				t.Fatal(err)
			}
			if client.disabled != test.wantDisabled {
				t.Fatalf("disabled = %v, want %v", client.disabled, test.wantDisabled)
			}
			if client.IsDisabled() != test.wantDisabled {
				t.Fatalf("IsDisabled() = %v, want %v", client.IsDisabled(), test.wantDisabled)
			}
			if (client.service == nil) != test.wantServiceNil {
				t.Fatalf("service nil = %v, want %v", client.service == nil, test.wantServiceNil)
			}
		})
	}
}

func TestNewClientDisablesWithoutDefaultBaseURL(t *testing.T) {
	setDefaultBaseURL(t, "")
	// An explicit user preference to enable telemetry cannot make a client
	// operational when its build does not contain a service endpoint.
	t.Setenv(telemetryDisabledEnv, "0")

	client, err := NewClient(&Config{})
	if err != nil {
		t.Fatal(err)
	}
	if !client.IsDisabled() {
		t.Fatal("client without a default base URL is enabled")
	}
	if client.service != nil {
		t.Fatal("client without a default base URL constructed a service")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() = %v", err)
	}
	if err := client.SendEvents(ctx, nil); err != nil {
		t.Fatalf("SendEvents() = %v", err)
	}
}

func TestDisabledClientDoesNothing(t *testing.T) {
	client := &Client{disabled: true}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() = %v", err)
	}
	if err := client.SendEvents(ctx, nil); err != nil {
		t.Fatalf("SendEvents() = %v", err)
	}
}

func TestSendEventsBoundaryValidation(t *testing.T) {
	client := &Client{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.SendEvents(ctx, []Event{{}}); !stderrors.Is(err, context.Canceled) {
		t.Fatalf("canceled SendEvents() = %v", err)
	}

	for _, events := range [][]Event{nil, make([]Event, maxBatchSize+1)} {
		if err := client.SendEvents(context.Background(), events); !stderrors.Is(err, telemetryerrors.ErrInvalidBatch) {
			t.Fatalf("SendEvents(%d events) = %v", len(events), err)
		}
	}
}

func stringPointer(value string) *string { return &value }

func setDefaultBaseURL(t *testing.T, value string) {
	t.Helper()
	original := defaultBaseURL
	defaultBaseURL = value
	t.Cleanup(func() { defaultBaseURL = original })
}

func TestNewEvent(t *testing.T) {
	sessionID := uuid.New()
	before := time.Now().UTC()
	event := NewEvent(
		sessionID,
		EventSourceCLI,
		OperationTestCreate,
		OperationOutcomeError,
		1500*time.Millisecond,
		OperationErrorCategoryValidation,
		OperationErrorCodeInvalidPayload,
		"v1.2.3",
	)
	after := time.Now().UTC()

	if event.ID.Compare(uuid.Nil()) == 0 {
		t.Fatal("event ID is zero")
	}
	if event.SessionID != sessionID {
		t.Fatalf("session ID = %s, want %s", event.SessionID, sessionID)
	}
	if event.SchemaVersion != SchemaVersion {
		t.Fatalf("schema version = %d, want %d", event.SchemaVersion, SchemaVersion)
	}
	if event.Timestamp.Location() != time.UTC || event.Timestamp.Before(before) || event.Timestamp.After(after) {
		t.Fatalf("timestamp = %s, want between %s and %s in UTC", event.Timestamp, before, after)
	}
	if event.Source != EventSourceCLI || event.Operation != OperationTestCreate || event.Outcome != OperationOutcomeError {
		t.Fatalf("operation fields = (%q, %q, %q)", event.Source, event.Operation, event.Outcome)
	}
	if event.DurationMS != 1500 {
		t.Fatalf("duration = %dms, want 1500ms", event.DurationMS)
	}
	if event.ErrorCategory != OperationErrorCategoryValidation || event.ErrorCode != OperationErrorCodeInvalidPayload {
		t.Fatalf("error fields = (%q, %q)", event.ErrorCategory, event.ErrorCode)
	}
	if event.OS != expectedOS() || event.Arch != expectedArch() {
		t.Fatalf("platform = (%q, %q), want (%q, %q)", event.OS, event.Arch, expectedOS(), expectedArch())
	}
	if event.Version != "v1.2.3" {
		t.Fatalf("version = %q", event.Version)
	}

	second := NewEvent(sessionID, EventSourceCLI, OperationTestCreate, OperationOutcomeSuccess, 0, "", "", "v1.2.3")
	if second.ID == event.ID {
		t.Fatal("NewEvent reused an event ID")
	}
}

func expectedOS() OS {
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

func expectedArch() Arch {
	switch runtime.GOARCH {
	case "amd64":
		return ArchAMD64
	case "arm64":
		return ArchARM64
	default:
		return ArchUnknown
	}
}
