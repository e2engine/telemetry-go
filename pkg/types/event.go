package types

import (
	"time"
	"uuid"
)

const SchemaVersion = 1

type Event struct {
	ID            uuid.UUID `json:"id"`
	SchemaVersion int       `json:"schema_version"`
	SessionID     uuid.UUID `json:"session_id"`
	Timestamp     time.Time `json:"timestamp"`

	Source     EventSource      `json:"source"`
	Operation  Operation        `json:"operation"`
	Outcome    OperationOutcome `json:"outcome"`
	DurationMS int64            `json:"duration_ms"`

	ErrorCategory OperationErrorCategory `json:"error_category,omitempty"`
	ErrorCode     OperationErrorCode     `json:"error_code,omitempty"`

	OS      OS     `json:"os"`
	Arch    Arch   `json:"arch"`
	Version string `json:"version"`
}

type EventSource string

const (
	EventSourceCLI EventSource = "cli"
	EventSourceMCP EventSource = "mcp"
)

type Operation string

const (
	OperationEnvironmentList   Operation = "environment.list"
	OperationEnvironmentGet    Operation = "environment.get"
	OperationEnvironmentCreate Operation = "environment.create"
	OperationEnvironmentDelete Operation = "environment.delete"
	OperationTestList          Operation = "test.list"
	OperationTestGet           Operation = "test.get"
	OperationTestCreate        Operation = "test.create"
	OperationTestDelete        Operation = "test.delete"
	OperationSuiteList         Operation = "suite.list"
	OperationSuiteGet          Operation = "suite.get"
	OperationSuiteCreate       Operation = "suite.create"
	OperationSuiteDelete       Operation = "suite.delete"
)

type OperationOutcome string

const (
	OperationOutcomeSuccess OperationOutcome = "success"
	OperationOutcomeError   OperationOutcome = "error"
)

type OS string

const (
	OSLinux   OS = "linux"
	OSDarwin  OS = "darwin"
	OSWindows OS = "windows"
	OSUnknown OS = "unknown"
)

type Arch string

const (
	ArchAMD64   Arch = "amd64"
	ArchARM64   Arch = "arm64"
	ArchUnknown Arch = "unknown"
)
