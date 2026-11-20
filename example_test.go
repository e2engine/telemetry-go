package telemetry_go_test

import (
	"context"
	"time"
	"uuid"

	telemetry "github.com/e2engine/telemetry-go"
)

func ExampleClient() {
	client, err := telemetry.NewClient(&telemetry.Config{})
	if err != nil {
		return
	}
	// One total budget bounds installation initialization and submission.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Initialize(ctx); err != nil {
		// Telemetry is best-effort. The application operation continues.
		return
	}
	// Create once per CLI invocation or MCP process, and reuse for its operations.
	sessionID := uuid.New()
	event := telemetry.NewEvent(
		sessionID,
		telemetry.EventSourceCLI,
		telemetry.OperationTestCreate,
		telemetry.OperationOutcomeSuccess,
		10*time.Millisecond,
		"",
		"",
		"v0.1.0",
	)
	// Synchronous and bounded, so short-lived commands can send before exit.
	// No automatic retry or background goroutine is needed.
	_ = client.SendEvents(ctx, []telemetry.Event{event})
}
