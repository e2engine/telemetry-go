package keys

import "github.com/ygrebnov/keys"

var event = keys.Factory(keys.WithSegments("event"))

var (
	Cause      = keys.New("cause")
	Operation  = keys.New("operation")
	Validation = keys.New("validation")
	HTTPStatus = keys.New("http_status")

	EventCount = event("count")
	EventLimit = event("limit")
)
