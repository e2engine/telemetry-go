package core

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ygrebnov/errorc"

	"github.com/e2engine/telemetry-go/internal/model"
	telemetryerrors "github.com/e2engine/telemetry-go/pkg/errors"
	"github.com/e2engine/telemetry-go/pkg/keys"
)

// Register explicitly provisions an installation.
func (s *Service) Register(ctx context.Context) (model.Installation, error) {
	resp, err := s.request(ctx, "/v1/installations", "", nil)
	if err != nil {
		return model.Installation{}, errors.Join(
			telemetryerrors.ErrRegisterInstallation,
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return model.Installation{}, errorc.With(
			telemetryerrors.ErrRegisterInstallation,
			errorc.String(keys.Cause, "unexpected response status code"),
			errorc.Int(keys.HTTPStatus, resp.StatusCode),
		)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxRegistrationBytes+1))
	if err != nil {
		return model.Installation{}, errors.Join(
			errorc.With(
				telemetryerrors.ErrRegisterInstallation,
				errorc.String(keys.Cause, "cannot read response body"),
			),
			err,
		)
	}

	if len(body) > maxRegistrationBytes {
		return model.Installation{}, errorc.With(
			telemetryerrors.ErrRegisterInstallation,
			errorc.String(keys.Cause, "response body exceeds maximum allowed size"),
		)
	}

	var registered model.Installation
	if json.Unmarshal(body, &registered) != nil || !registered.IsValid() {
		return model.Installation{}, errors.Join(
			telemetryerrors.ErrRegisterInstallation,
			telemetryerrors.ErrInvalidInstallationData,
		)
	}

	return registered, nil
}
