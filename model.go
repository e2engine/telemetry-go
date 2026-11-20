package telemetry_go

import "github.com/e2engine/telemetry-go/pkg/types"

type (
	Event                  = types.Event
	EventSource            = types.EventSource
	Operation              = types.Operation
	OperationOutcome       = types.OperationOutcome
	OS                     = types.OS
	Arch                   = types.Arch
	OperationErrorCategory = types.OperationErrorCategory
	OperationErrorCode     = types.OperationErrorCode
)

const (
	SchemaVersion                                     = types.SchemaVersion
	EventSourceCLI                                    = types.EventSourceCLI
	EventSourceMCP                                    = types.EventSourceMCP
	OperationEnvironmentList                          = types.OperationEnvironmentList
	OperationEnvironmentGet                           = types.OperationEnvironmentGet
	OperationEnvironmentCreate                        = types.OperationEnvironmentCreate
	OperationEnvironmentDelete                        = types.OperationEnvironmentDelete
	OperationTestList                                 = types.OperationTestList
	OperationTestGet                                  = types.OperationTestGet
	OperationTestCreate                               = types.OperationTestCreate
	OperationTestDelete                               = types.OperationTestDelete
	OperationSuiteList                                = types.OperationSuiteList
	OperationSuiteGet                                 = types.OperationSuiteGet
	OperationSuiteCreate                              = types.OperationSuiteCreate
	OperationSuiteDelete                              = types.OperationSuiteDelete
	OperationOutcomeSuccess                           = types.OperationOutcomeSuccess
	OperationOutcomeError                             = types.OperationOutcomeError
	OSLinux                                           = types.OSLinux
	OSDarwin                                          = types.OSDarwin
	OSWindows                                         = types.OSWindows
	OSUnknown                                         = types.OSUnknown
	ArchAMD64                                         = types.ArchAMD64
	ArchARM64                                         = types.ArchARM64
	ArchUnknown                                       = types.ArchUnknown
	OperationErrorCategoryValidation                  = types.OperationErrorCategoryValidation
	OperationErrorCategoryAuthentication              = types.OperationErrorCategoryAuthentication
	OperationErrorCategoryAuthorization               = types.OperationErrorCategoryAuthorization
	OperationErrorCategoryNotFound                    = types.OperationErrorCategoryNotFound
	OperationErrorCategoryConflict                    = types.OperationErrorCategoryConflict
	OperationErrorCategoryNetwork                     = types.OperationErrorCategoryNetwork
	OperationErrorCategoryTimeout                     = types.OperationErrorCategoryTimeout
	OperationErrorCategoryUnavailable                 = types.OperationErrorCategoryUnavailable
	OperationErrorCategoryInternal                    = types.OperationErrorCategoryInternal
	OperationErrorCategoryUnknown                     = types.OperationErrorCategoryUnknown
	OperationErrorCodeResourceNotFound                = types.OperationErrorCodeResourceNotFound
	OperationErrorCodeResourceConflict                = types.OperationErrorCodeResourceConflict
	OperationErrorCodeResourceAlreadyExists           = types.OperationErrorCodeResourceAlreadyExists
	OperationErrorCodeInvalidPayload                  = types.OperationErrorCodeInvalidPayload
	OperationErrorCodeInvalidPageTraversal            = types.OperationErrorCodeInvalidPageTraversal
	OperationErrorCodeInvalidPageLimit                = types.OperationErrorCodeInvalidPageLimit
	OperationErrorCodeInvalidExecutionStatus          = types.OperationErrorCodeInvalidExecutionStatus
	OperationErrorCodeInvalidType                     = types.OperationErrorCodeInvalidType
	OperationErrorCodeTooManyResults                  = types.OperationErrorCodeTooManyResults
	OperationErrorCodeInvalidTransportConfig          = types.OperationErrorCodeInvalidTransportConfig
	OperationErrorCodeInvalidTransportMessage         = types.OperationErrorCodeInvalidTransportMessage
	OperationErrorCodeInvalidDatabasePath             = types.OperationErrorCodeInvalidDatabasePath
	OperationErrorCodeInvalidExecutionStateTransition = types.OperationErrorCodeInvalidExecutionStateTransition
	OperationErrorCodeDatabaseConnectionFailed        = types.OperationErrorCodeDatabaseConnectionFailed
	OperationErrorCodeDatabaseConfigurationFailed     = types.OperationErrorCodeDatabaseConfigurationFailed
	OperationErrorCodeDatabaseMigrationFailed         = types.OperationErrorCodeDatabaseMigrationFailed
	OperationErrorCodeDatabaseWALFailed               = types.OperationErrorCodeDatabaseWALFailed
	OperationErrorCodeRunnerNotStarted                = types.OperationErrorCodeRunnerNotStarted
	OperationErrorCodeAuthenticationFailed            = types.OperationErrorCodeAuthenticationFailed
	OperationErrorCodePermissionDenied                = types.OperationErrorCodePermissionDenied
	OperationErrorCodeNetworkError                    = types.OperationErrorCodeNetworkError
	OperationErrorCodeTimeout                         = types.OperationErrorCodeTimeout
	OperationErrorCodeServiceUnavailable              = types.OperationErrorCodeServiceUnavailable
	OperationErrorCodeInternalError                   = types.OperationErrorCodeInternalError
	OperationErrorCodeUnknownError                    = types.OperationErrorCodeUnknownError
)
