// Package errors defines telemetry error classifications for use with errors.Is.
package errors

import "github.com/ygrebnov/errorc"

var (
	ErrNilConfig                      = errorc.New("nil configuration")
	ErrInvalidConfig                  = errorc.New("invalid configuration")
	ErrInvalidURL                     = errorc.New("invalid URL")
	ErrInvalidInstallationData        = errorc.New("invalid installation data")
	ErrInvalidInstallation            = errorc.New("invalid installation file")
	ErrUnsupportedInstallationVersion = errorc.New("unsupported installation file version")
	ErrInsecureInstallation           = errorc.New("installation permissions are not private")
	ErrLoadInstallation               = errorc.New("cannot load installation")
	ErrSaveInstallation               = errorc.New("cannot save installation")
	ErrInvalidBatch                   = errorc.New("invalid event batch")
	ErrRegisterInstallation           = errorc.New("cannot register installation")
	ErrSendEvents                     = errorc.New("cannot send events")
	ErrConstructRequest               = errorc.New("cannot construct HTTP request")
	ErrDoRequest                      = errorc.New("cannot do HTTP request")
	ErrInvalidToken                   = errorc.New("invalid installation token")
	ErrResolveInstallationPath        = errorc.New("cannot resolve installation path")
)
