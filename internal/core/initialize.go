package core

import (
	"context"
	"errors"
	"os"

	"github.com/e2engine/telemetry-go/internal/fs"
	telemetryerrors "github.com/e2engine/telemetry-go/pkg/errors"
)

// Initialize loads or creates installation data at most once.
func (s *Service) Initialize(ctx context.Context) error {
	s.initializeOnce.Do(func() {
		s.initializeErr = s.initialize(ctx)
	})

	return s.initializeErr
}

func (s *Service) initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	path, err := fs.DefaultInstallationPath()
	if err != nil {
		return err
	}

	store := fs.NewInstallationStore(path)

	installation, err := store.Load()
	if err != nil &&
		(errors.Is(err, os.ErrNotExist) ||
			errors.Is(err, telemetryerrors.ErrInvalidInstallation)) {
		// attempt to register a new installation
		installation, err = s.Register(ctx)
		if err != nil {
			return err
		}

		err = store.Save(installation)
	}
	if err != nil {
		return err
	}

	if errCtx := ctx.Err(); errCtx != nil {
		return errCtx
	}

	s.token = installation.Token
	return nil
}
