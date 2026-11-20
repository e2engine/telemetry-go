package core

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/e2engine/telemetry-go/internal/fs"
	"github.com/e2engine/telemetry-go/internal/model"
	telemetryerrors "github.com/e2engine/telemetry-go/pkg/errors"
	"github.com/e2engine/telemetry-go/pkg/types"
)

var coreTestInstallation = coreValidInstallation("A")

func coreValidInstallation(character string) model.Installation {
	id := strings.Repeat(character, 21) + "A"
	secret := strings.Repeat(character, 42) + "A"
	return model.Installation{ID: id, Token: "et_" + id + "." + secret}
}

func newCoreTestService(t *testing.T, handler http.HandlerFunc) (*Service, *atomic.Int32) {
	t.Helper()
	calls := new(atomic.Int32)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		handler(writer, request)
	}))
	t.Cleanup(server.Close)
	return newService(server.URL, getHTTPClient(250*time.Millisecond)), calls
}

func setUserConfigRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	switch runtime.GOOS {
	case "darwin":
		t.Setenv("HOME", root)
	case "windows":
		t.Setenv("AppData", root)
	default:
		t.Setenv("XDG_CONFIG_HOME", root)
	}
	return root
}

func TestRegister(t *testing.T) {
	service, calls := newCoreTestService(t, func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/v1/installations" {
			t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Authorization") != "" {
			t.Error("registration included authorization")
		}
		writer.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(writer).Encode(coreTestInstallation); err != nil {
			t.Error(err)
		}
	})

	got, err := service.Register(context.Background())
	if err != nil || got != coreTestInstallation {
		t.Fatalf("Register() = (%v, %v)", got, err)
	}
	if calls.Load() != 1 {
		t.Fatalf("requests = %d", calls.Load())
	}
}

func TestRegisterFailures(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "status", status: http.StatusServiceUnavailable, body: "sensitive response"},
		{name: "malformed", status: http.StatusCreated, body: "{"},
		{name: "empty", status: http.StatusCreated, body: "{}"},
		{name: "trailing", status: http.StatusCreated, body: "{} {}"},
		{name: "oversized", status: http.StatusCreated, body: strings.Repeat("x", maxRegistrationBytes+1)},
		{name: "mismatch", status: http.StatusCreated, body: `{"id":"wrong","token":"secret"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, calls := newCoreTestService(t, func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(test.status)
				_, _ = io.WriteString(writer, test.body)
			})
			got, err := service.Register(context.Background())
			if got != (model.Installation{}) || !errors.Is(err, telemetryerrors.ErrRegisterInstallation) {
				t.Fatalf("Register() = (%v, %v)", got, err)
			}
			if calls.Load() != 1 {
				t.Fatalf("requests = %d", calls.Load())
			}
			if strings.Contains(err.Error(), test.body) {
				t.Fatal("error exposed response body")
			}
		})
	}
}

func TestInitializeRegistersPersistsAndReuses(t *testing.T) {
	setUserConfigRoot(t)
	service, calls := newCoreTestService(t, func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(coreTestInstallation)
	})

	if err := service.Initialize(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := service.Initialize(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("registrations = %d", calls.Load())
	}

	second, _ := newCoreTestService(t, func(http.ResponseWriter, *http.Request) {
		t.Error("registered instead of reusing installation")
	})
	if err := second.Initialize(context.Background()); err != nil {
		t.Fatal(err)
	}
	if second.token != coreTestInstallation.Token {
		t.Fatal("reused wrong token")
	}
}

func TestInitializeFailureIsCached(t *testing.T) {
	setUserConfigRoot(t)
	service, calls := newCoreTestService(t, func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusServiceUnavailable)
	})

	first := service.Initialize(context.Background())
	second := service.Initialize(context.Background())
	if !errors.Is(first, telemetryerrors.ErrRegisterInstallation) {
		t.Fatalf("first Initialize() = %v", first)
	}
	if !errors.Is(second, telemetryerrors.ErrRegisterInstallation) {
		t.Fatalf("second Initialize() = %v", second)
	}
	if first.Error() != second.Error() {
		t.Fatalf("cached error changed: %q != %q", first, second)
	}
	if calls.Load() != 1 {
		t.Fatalf("registrations = %d", calls.Load())
	}
}

func TestConcurrentInitializeRunsOnce(t *testing.T) {
	setUserConfigRoot(t)
	registrationStarted := make(chan struct{})
	releaseRegistration := make(chan struct{})
	service, calls := newCoreTestService(t, func(writer http.ResponseWriter, _ *http.Request) {
		close(registrationStarted)
		<-releaseRegistration
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(coreTestInstallation)
	})

	const callers = 8
	results := make(chan error, callers)
	for range callers {
		go func() { results <- service.Initialize(context.Background()) }()
	}
	select {
	case <-registrationStarted:
	case <-time.After(time.Second):
		t.Fatal("registration did not start")
	}
	close(releaseRegistration)

	for range callers {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("registrations = %d", calls.Load())
	}
}

func TestInitializeRecoveryPolicy(t *testing.T) {
	t.Run("malformed version one", func(t *testing.T) {
		setUserConfigRoot(t)
		path, err := fs.DefaultInstallationPath()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(`{"version":1,"token":"invalid"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		service, calls := newCoreTestService(t, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(writer).Encode(coreTestInstallation)
		})
		if err := service.Initialize(context.Background()); err != nil {
			t.Fatal(err)
		}
		if calls.Load() != 1 {
			t.Fatalf("registrations = %d", calls.Load())
		}
	})

	t.Run("future version", func(t *testing.T) {
		setUserConfigRoot(t)
		path, err := fs.DefaultInstallationPath()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		const contents = `{"version":2,"future":true}`
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
		service, calls := newCoreTestService(t, func(http.ResponseWriter, *http.Request) {
			t.Error("registered over future installation")
		})
		if err := service.Initialize(context.Background()); !errors.Is(err, telemetryerrors.ErrUnsupportedInstallationVersion) {
			t.Fatalf("Initialize() = %v", err)
		}
		if calls.Load() != 0 {
			t.Fatalf("requests = %d", calls.Load())
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != contents {
			t.Fatalf("future file changed: %q, %v", data, err)
		}
	})
}

func TestSendEvents(t *testing.T) {
	event := types.Event{SchemaVersion: types.SchemaVersion, Source: types.EventSourceCLI, Version: "v0.1.0"}
	service, calls := newCoreTestService(t, func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/v1/events" {
			t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer "+coreTestInstallation.Token {
			t.Error("missing bearer authentication")
		}
		var body sendEventsBody
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil || len(body.Events) != 1 || body.Events[0] != event {
			t.Errorf("invalid body: %#v, %v", body, err)
		}
		writer.WriteHeader(http.StatusNoContent)
	})
	service.token = coreTestInstallation.Token

	if err := service.SendEvents(context.Background(), []types.Event{event}); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("requests = %d", calls.Load())
	}
}

func TestSendEventsFailuresAreNotRetried(t *testing.T) {
	for _, status := range []int{http.StatusTemporaryRedirect, http.StatusBadRequest, http.StatusUnauthorized, http.StatusRequestEntityTooLarge, http.StatusTooManyRequests, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			service, calls := newCoreTestService(t, func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Location", "/redirected")
				writer.WriteHeader(status)
				_, _ = io.WriteString(writer, "sensitive response")
			})
			service.token = coreTestInstallation.Token
			err := service.SendEvents(context.Background(), []types.Event{{}})
			if !errors.Is(err, telemetryerrors.ErrSendEvents) {
				t.Fatalf("SendEvents() = %v", err)
			}
			if calls.Load() != 1 {
				t.Fatalf("requests = %d", calls.Load())
			}
			if strings.Contains(err.Error(), "sensitive response") || strings.Contains(err.Error(), coreTestInstallation.Token) {
				t.Fatal("error exposed sensitive data")
			}
		})
	}
}

func TestSendEventsValidationAndCancellation(t *testing.T) {
	service, calls := newCoreTestService(t, func(http.ResponseWriter, *http.Request) {
		t.Error("unexpected request")
	})
	service.token = coreTestInstallation.Token

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := service.SendEvents(ctx, []types.Event{{}}); !errors.Is(err, context.Canceled) {
		t.Fatalf("SendEvents(canceled) = %v", err)
	}
	large := types.Event{Version: strings.Repeat("x", maxEventsBytes)}
	if err := service.SendEvents(context.Background(), []types.Event{large}); !errors.Is(err, telemetryerrors.ErrSendEvents) {
		t.Fatalf("SendEvents(large) = %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("requests = %d", calls.Load())
	}
}

func TestSendWaitsForConcurrentInitialization(t *testing.T) {
	setUserConfigRoot(t)
	registrationStarted := make(chan struct{})
	releaseRegistration := make(chan struct{})
	service, calls := newCoreTestService(t, func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/v1/installations":
			close(registrationStarted)
			<-releaseRegistration
			writer.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(writer).Encode(coreTestInstallation)
		case "/v1/events":
			writer.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected path %q", request.URL.Path)
		}
	})

	initializeResult := make(chan error, 1)
	go func() { initializeResult <- service.Initialize(context.Background()) }()
	select {
	case <-registrationStarted:
	case <-time.After(time.Second):
		t.Fatal("registration did not start")
	}

	sendResult := make(chan error, 1)
	go func() { sendResult <- service.SendEvents(context.Background(), []types.Event{{}}) }()
	close(releaseRegistration)

	for name, result := range map[string]<-chan error{"initialize": initializeResult, "send": sendResult} {
		select {
		case err := <-result:
			if err != nil {
				t.Fatalf("%s = %v", name, err)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("%s deadlocked", name)
		}
	}
	if calls.Load() != 2 {
		t.Fatalf("requests = %d", calls.Load())
	}
}

func TestRequestTimeoutPreservesCause(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
	}))
	t.Cleanup(server.Close)
	service := newService(server.URL, getHTTPClient(20*time.Millisecond))

	_, err := service.Register(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Register() = %v", err)
	}
}
