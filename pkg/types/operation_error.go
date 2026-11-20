package types

type OperationErrorCategory string

const (
	// OperationErrorCategoryValidation represents invalid specification or input errors.
	OperationErrorCategoryValidation OperationErrorCategory = "validation"

	// OperationErrorCategoryAuthentication represents missing or invalid credentials errors.
	OperationErrorCategoryAuthentication OperationErrorCategory = "authentication"

	// OperationErrorCategoryAuthorization represents insufficient permissions errors.
	OperationErrorCategoryAuthorization OperationErrorCategory = "authorization"

	// OperationErrorCategoryNotFound represents requested resource does not exist errors.
	OperationErrorCategoryNotFound OperationErrorCategory = "not_found"

	// OperationErrorCategoryConflict represents resource already exists or version conflict errors.
	OperationErrorCategoryConflict OperationErrorCategory = "conflict"

	// OperationErrorCategoryNetwork represents connection or transport failure errors.
	OperationErrorCategoryNetwork OperationErrorCategory = "network"

	// OperationErrorCategoryTimeout represents operation exceeded its deadline errors.
	OperationErrorCategoryTimeout OperationErrorCategory = "timeout"

	// OperationErrorCategoryUnavailable represents required service temporarily unavailable errors.
	OperationErrorCategoryUnavailable OperationErrorCategory = "unavailable"

	// OperationErrorCategoryInternal represents unexpected internal failure errors.
	OperationErrorCategoryInternal OperationErrorCategory = "internal"

	// OperationErrorCategoryUnknown represents unknown or uncategorized errors.
	OperationErrorCategoryUnknown OperationErrorCategory = "unknown"
)

type OperationErrorCode string

const (
	// Resource errors.
	OperationErrorCodeResourceNotFound      OperationErrorCode = "resource_not_found"
	OperationErrorCodeResourceConflict      OperationErrorCode = "resource_conflict"
	OperationErrorCodeResourceAlreadyExists OperationErrorCode = "resource_already_exists"

	// Validation errors.
	OperationErrorCodeInvalidPayload          OperationErrorCode = "invalid_payload"
	OperationErrorCodeInvalidPageTraversal    OperationErrorCode = "invalid_page_traversal"
	OperationErrorCodeInvalidPageLimit        OperationErrorCode = "invalid_page_limit"
	OperationErrorCodeInvalidExecutionStatus  OperationErrorCode = "invalid_execution_status"
	OperationErrorCodeInvalidType             OperationErrorCode = "invalid_type"
	OperationErrorCodeTooManyResults          OperationErrorCode = "too_many_results"
	OperationErrorCodeInvalidTransportConfig  OperationErrorCode = "invalid_transport_config"
	OperationErrorCodeInvalidTransportMessage OperationErrorCode = "invalid_transport_message"
	OperationErrorCodeInvalidDatabasePath     OperationErrorCode = "invalid_database_path"

	// Execution errors.
	OperationErrorCodeInvalidExecutionStateTransition OperationErrorCode = "invalid_execution_state_transition"

	// Database errors.
	OperationErrorCodeDatabaseConnectionFailed    OperationErrorCode = "database_connection_failed"
	OperationErrorCodeDatabaseConfigurationFailed OperationErrorCode = "database_configuration_failed"
	OperationErrorCodeDatabaseMigrationFailed     OperationErrorCode = "database_migration_failed"
	OperationErrorCodeDatabaseWALFailed           OperationErrorCode = "database_wal_failed"

	// Runner errors.
	OperationErrorCodeRunnerNotStarted OperationErrorCode = "runner_not_started"

	// General errors.
	OperationErrorCodeAuthenticationFailed OperationErrorCode = "authentication_failed"
	OperationErrorCodePermissionDenied     OperationErrorCode = "permission_denied"
	OperationErrorCodeNetworkError         OperationErrorCode = "network_error"
	OperationErrorCodeTimeout              OperationErrorCode = "timeout"
	OperationErrorCodeServiceUnavailable   OperationErrorCode = "service_unavailable"
	OperationErrorCodeInternalError        OperationErrorCode = "internal_error"
	OperationErrorCodeUnknownError         OperationErrorCode = "unknown_error"
)
