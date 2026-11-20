package fs

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/ygrebnov/errorc"

	"github.com/e2engine/telemetry-go/internal/model"
	telemetryerrors "github.com/e2engine/telemetry-go/pkg/errors"
	"github.com/e2engine/telemetry-go/pkg/keys"
)

const installationFileVersion = 1
const maxInstallationBytes = 4096

// DefaultInstallationPath returns OS-specific default path for the installation file,
// which is used to persist the installation ID and token.
func DefaultInstallationPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", errors.Join(
			telemetryerrors.ErrResolveInstallationPath,
			err,
		)
	}
	return filepath.Join(dir, "e2engine", "installation.json"), nil
}

// InstallationStore performs filesystem operations to load and save the installation file.
type InstallationStore struct {
	path string
}

func NewInstallationStore(path string) *InstallationStore {
	return &InstallationStore{
		path: path,
	}
}

type installationFile struct {
	Version int    `json:"version"`
	ID      string `json:"installation_id"`
	Token   string `json:"token"`
}

func (s *InstallationStore) root(create bool) (*os.Root, error) {
	if s.path == "" {
		return nil, errorc.With(
			telemetryerrors.ErrInvalidConfig,
			errorc.String(keys.Validation, "empty installation path"),
		)
	}

	dir := filepath.Dir(s.path)
	if create {
		if err := prepareDirectory(dir); err != nil {
			return nil, err
		}
	}

	info, err := os.Lstat(dir)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, telemetryerrors.ErrInsecureInstallation
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	file, err := root.Open(".")
	if err == nil {
		err = checkPrivate(file)
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}
	if err != nil {
		_ = root.Close()
		return nil, err
	}

	return root, nil
}

func safeTarget(root *os.Root, name string) error {
	info, err := root.Lstat(name)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return telemetryerrors.ErrInsecureInstallation
	}
	return nil
}

// Load returns os.ErrNotExist for an absent installation, typed sentinel errors
// for malformed/unsupported files, and original filesystem errors for I/O failures.
func (s *InstallationStore) Load() (_ model.Installation, err error) {
	defer func() {
		switch {
		case err == nil:
		case errors.Is(err, telemetryerrors.ErrInvalidInstallation),
			errors.Is(err, telemetryerrors.ErrUnsupportedInstallationVersion),
			errors.Is(err, telemetryerrors.ErrInsecureInstallation):
			err = errorc.With(
				err,
				errorc.String(keys.Operation, "load installation"),
			)
		default:
			err = errors.Join(
				errorc.With(
					telemetryerrors.ErrLoadInstallation,
					errorc.String(keys.Operation, "load installation"),
				),
				err,
			)
		}
	}()
	root, err := s.root(false)
	if err != nil {
		return model.Installation{}, err
	}
	defer root.Close()
	name := filepath.Base(s.path)
	if targetErr := safeTarget(root, name); targetErr != nil {
		return model.Installation{}, targetErr
	}
	file, err := root.Open(name)
	if err != nil {
		return model.Installation{}, err
	}
	defer file.Close()
	if permissionErr := checkPrivate(file); permissionErr != nil {
		return model.Installation{}, permissionErr
	}
	data, err := io.ReadAll(io.LimitReader(file, maxInstallationBytes+1))
	if err != nil {
		return model.Installation{}, err
	}
	if len(data) > maxInstallationBytes {
		return model.Installation{}, telemetryerrors.ErrInvalidInstallation
	}
	// Inspect version before validating fields so newer formats are never replaced.
	var version struct {
		Version int `json:"version"`
	}
	if json.Unmarshal(data, &version) != nil {
		return model.Installation{}, telemetryerrors.ErrInvalidInstallation
	}
	if version.Version > installationFileVersion {
		return model.Installation{}, errorc.With(
			telemetryerrors.ErrUnsupportedInstallationVersion,
			errorc.Int("file.version", version.Version),
		)
	}
	var stored installationFile
	if json.Unmarshal(data, &stored) != nil || stored.Version != installationFileVersion {
		return model.Installation{}, telemetryerrors.ErrInvalidInstallation
	}
	installation := model.Installation{ID: stored.ID, Token: stored.Token}
	if !installation.IsValid() {
		return model.Installation{}, telemetryerrors.ErrInvalidInstallation
	}
	return installation, nil
}

// Save writes privately to a unique same-directory file, syncs and closes it,
// then atomically replaces the destination. Concurrent writers may replace each
// other. Unsupported native replacement operations fail without a copy fallback.
func (s *InstallationStore) Save(value model.Installation) (err error) {
	defer func() {
		if err != nil {
			err = errors.Join(
				errorc.With(
					telemetryerrors.ErrSaveInstallation,
					errorc.String(keys.Operation, "save installation"),
				),
				err,
			)
		}
	}()
	if !value.IsValid() {
		return telemetryerrors.ErrInvalidInstallationData
	}
	root, err := s.root(true)
	if err != nil {
		return err
	}
	defer root.Close()
	name := filepath.Base(s.path)
	if targetErr := safeTarget(root, name); targetErr != nil && !errors.Is(targetErr, os.ErrNotExist) {
		return targetErr
	}
	// O_EXCL plus a random name prevents writers from sharing a temporary file.
	tempName := ".installation-" + rand.Text() + ".tmp"
	file, err := root.OpenFile(tempName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close(); _ = root.Remove(tempName) }()
	if permissionErr := makePrivate(file); permissionErr != nil {
		return permissionErr
	}
	data, err := json.Marshal(installationFile{
		Version: installationFileVersion, ID: value.ID, Token: value.Token,
	})
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return replaceInstallation(root, filepath.Dir(s.path), tempName, name)
}
