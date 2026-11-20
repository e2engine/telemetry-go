# E2Engine Telemetry for Go

E2Engine telemetry client. Small, synchronous, and best-effort.
The existing `Event` model and predefined operation and error constants are the
wire schema. Events contain no raw diagnostic messages, stack traces, file paths,
project identifiers, or arbitrary payloads.

Part of [E2Engine](https://e2engine.dev), an open-source platform for declarative end-to-end testing.

E2Engine project on GitHub: [E2Engine on GitHub](https://github.com/e2engine).

## Configuration and opt-out

`Config` is intended to be embedded in the application configuration. Telemetry
is enabled by default and has one user setting:

```yaml
telemetry:
  disabled: false
```

`NewClient` checks `E2ENGINE_TELEMETRY_DISABLED` again when it snapshots the
configuration. An unset or empty value uses `Config.Disabled`. `0` and `false`
force telemetry on. Any other non-empty value disables telemetry. Values are
exact and case-sensitive.

A disabled client constructs no core service. Its `Initialize` and `SendEvents`
methods return without credential-file or network access. Disabling telemetry
does not delete previously saved installation credentials.

## Client lifecycle

```go
client, err := telemetry.NewClient(&telemetry.Config{})
if err != nil {
    // Continue without telemetry.
}

if client != nil {
    if err := client.Initialize(ctx); err != nil {
        // Continue without telemetry.
    }
}
```

`Initialize` loads the shared installation identity. If it is missing or contains
malformed version-1 data, the client calls `POST /v1/installations` and saves the
returned identity atomically. Initialization is idempotent within a process.

`SendEvents` submits one authenticated batch synchronously. Batches must contain
1–25 events and encode to at most 64 KiB. The client uses a private two-second
HTTP timeout, respects earlier context cancellation and deadlines, follows no
redirects, and makes no automatic retries. It has no persistent queue, worker,
daemon, or background goroutine.

HTTP 204 means that the telemetry service structurally accepted the request for
best-effort processing. It does not mean that every event was valid or persisted.
The service can filter individual invalid events and can lose events on a storage
failure.

Client methods return errors for optional diagnostics. CLI and MCP must not let a
telemetry error change the result of their primary operation. Returned errors are
classifiable with `errors.Is` and the sentinels in `pkg/errors`.

Create one session UUID per CLI invocation or MCP process lifetime and a distinct
event UUID for every operation. Use `EventSourceCLI` or `EventSourceMCP` and only
the predefined error categories and codes.

## Shared installation identity

CLI and MCP use the same file for an OS user profile:

```text
filepath.Join(os.UserConfigDir(), "e2engine", "installation.json")
```

The file is separate from the E2Engine database:

```json
{"version":1,"installation_id":"<server-issued ID>","token":"et_<ID>.<secret>"}
```

- Missing or malformed version-1 data is registered and replaced.
- A valid file is reused without registration.
- An unsupported future version is preserved and returns an error.
- Permission, security, and other I/O failures return an error without registration.
- Simultaneous initial registrations by separate processes may create different
  identities. The last atomic replacement wins.

Files are size-bounded and installation data is validated against the token
format. Nonregular files and symlinks are rejected. Unix directories and files
are created with `0700` and `0600` permissions. Existing overly broad permissions
are rejected. Windows uses a protected ACL for the current user, SYSTEM, and
administrators.

Saves use an exclusive random temporary file in the destination directory,
followed by write, sync, close, and native atomic replacement. There is no
cross-process lock.

## Verification

Unit and integration tests use `httptest` and temporary configuration directories.
They require no live AWS resources. Run tests separately from local end-to-end
verification with telemetry-service and DynamoDB Local. Do not publish this module
until that end-to-end flow passes.

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).