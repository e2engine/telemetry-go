package core

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ygrebnov/errorc"

	telemetryerrors "github.com/e2engine/telemetry-go/pkg/errors"
	"github.com/e2engine/telemetry-go/pkg/keys"
	"github.com/e2engine/telemetry-go/pkg/types"
)

func (s *Service) SendEvents(ctx context.Context, events []types.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	body, err := s.encodeEvents(events)
	if err != nil {
		return err
	}

	s.mu.Lock()
	token := s.token
	s.mu.Unlock()

	if token == "" {
		return errors.Join(
			telemetryerrors.ErrSendEvents,
			telemetryerrors.ErrInvalidToken,
		)
	}

	return s.sendEvents(ctx, body, token)
}

type sendEventsBody struct {
	Events []types.Event `json:"events"`
}

func (s *Service) encodeEvents(events []types.Event) ([]byte, error) {
	body := sendEventsBody{Events: events}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Join(
			telemetryerrors.ErrSendEvents,
			err,
		)
	}

	if len(b) > maxEventsBytes {
		return nil, errorc.With(
			telemetryerrors.ErrSendEvents,
			errorc.String(keys.Cause, "body size exceeds maximum allowed"),
		)
	}

	return b, nil
}

// sendEvents submits a batch encoded by encodeEvents with installation authentication.
func (s *Service) sendEvents(ctx context.Context, body []byte, token string) error {
	resp, err := s.request(ctx, "/v1/events", token, body)
	if err != nil {
		return errors.Join(
			telemetryerrors.ErrSendEvents,
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errorc.With(
			telemetryerrors.ErrSendEvents,
			errorc.String(keys.Cause, "unexpected response status code"),
			errorc.Int(keys.HTTPStatus, resp.StatusCode),
		)
	}

	return nil
}
