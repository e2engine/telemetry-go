package fs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/e2engine/telemetry-go/internal/model"
	telemetryerrors "github.com/e2engine/telemetry-go/pkg/errors"
)

var testInstallation = validInstallation("A")

func validInstallation(character string) model.Installation {
	id := strings.Repeat(character, 21) + "A"
	secret := strings.Repeat(character, 42) + "A"
	return model.Installation{ID: id, Token: "et_" + id + "." + secret}
}

func privateTestDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "profile")
	if err := prepareDirectory(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestInstallationStoreSaveLoadAndReplace(t *testing.T) {
	path := filepath.Join(privateTestDir(t), "installation.json")
	store := NewInstallationStore(path)
	if err := store.Save(testInstallation); err != nil {
		t.Fatal(err)
	}
	if got, err := store.Load(); err != nil || got != testInstallation {
		t.Fatalf("Load() = (%v, %v)", got, err)
	}

	invalid := model.Installation{ID: "invalid", Token: "private-secret"}
	if err := store.Save(invalid); !errors.Is(err, telemetryerrors.ErrInvalidInstallationData) ||
		!errors.Is(err, telemetryerrors.ErrSaveInstallation) {
		t.Fatalf("Save(invalid) = %v", err)
	}
	if got, err := store.Load(); err != nil || got != testInstallation {
		t.Fatalf("invalid save changed installation: (%v, %v)", got, err)
	}

	other := validInstallation("B")
	if err := store.Save(other); err != nil {
		t.Fatal(err)
	}
	if got, err := store.Load(); err != nil || got != other {
		t.Fatalf("replacement = (%v, %v)", got, err)
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "installation.json" {
		t.Fatalf("temporary files remain: %v", entries)
	}
}

func TestInstallationStoreConcurrentReplacement(t *testing.T) {
	path := filepath.Join(privateTestDir(t), "installation.json")
	store := NewInstallationStore(path)
	installations := []model.Installation{validInstallation("A"), validInstallation("B")}

	const writers = 6
	var wait sync.WaitGroup
	failures := make(chan error, writers)
	for writer := 0; writer < writers; writer++ {
		wait.Add(1)
		go func(value model.Installation) {
			defer wait.Done()
			for range 10 {
				if err := store.Save(value); err != nil {
					failures <- err
					return
				}
				got, err := NewInstallationStore(path).Load()
				if err != nil || (got != installations[0] && got != installations[1]) {
					failures <- fmt.Errorf("read incomplete installation: %w", err)
					return
				}
			}
		}(installations[writer%len(installations)])
	}
	wait.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
}

func TestInstallationStoreLoadFailures(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		path := filepath.Join(privateTestDir(t), "installation.json")
		_, err := NewInstallationStore(path).Load()
		if !errors.Is(err, os.ErrNotExist) || !errors.Is(err, telemetryerrors.ErrLoadInstallation) {
			t.Fatalf("Load() = %v", err)
		}
	})

	invalidFiles := []string{
		"",
		"null",
		"[]",
		`{}`,
		`{"version":0}`,
		`{"version":1} {}`,
		strings.Repeat("x", maxInstallationBytes+1),
		`{"version":1,"installation_id":"wrong","token":"` + testInstallation.Token + `"}`,
	}
	for index, data := range invalidFiles {
		t.Run(fmt.Sprintf("invalid-%d", index), func(t *testing.T) {
			path := filepath.Join(privateTestDir(t), "installation.json")
			if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := NewInstallationStore(path).Load(); !errors.Is(err, telemetryerrors.ErrInvalidInstallation) {
				t.Fatalf("Load() = %v", err)
			}
		})
	}
}

func TestInstallationStorePreservesFutureVersion(t *testing.T) {
	path := filepath.Join(privateTestDir(t), "installation.json")
	const contents = `{"version":2,"future":true}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewInstallationStore(path).Load(); !errors.Is(err, telemetryerrors.ErrUnsupportedInstallationVersion) {
		t.Fatalf("Load() = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != contents {
		t.Fatalf("future file changed: %q, %v", data, err)
	}
}

func TestInstallationStoreRejectsUnsafeTargets(t *testing.T) {
	for _, test := range []struct {
		name  string
		setup func(string) error
	}{
		{name: "directory", setup: func(path string) error { return os.Mkdir(path, 0o700) }},
		{name: "symlink", setup: func(path string) error { return os.Symlink(path+".missing", path) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			if runtime.GOOS == "windows" && test.name == "symlink" {
				t.Skip("symlink privileges differ on Windows")
			}
			path := filepath.Join(privateTestDir(t), "installation.json")
			if err := test.setup(path); err != nil {
				t.Fatal(err)
			}
			if _, err := NewInstallationStore(path).Load(); !errors.Is(err, telemetryerrors.ErrInsecureInstallation) {
				t.Fatalf("Load() = %v", err)
			}
		})
	}
}

func TestInstallationPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("covered by Windows ACL tests")
	}
	path := filepath.Join(privateTestDir(t), "installation.json")
	store := NewInstallationStore(path)
	if err := store.Save(testInstallation); err != nil {
		t.Fatal(err)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Mode().Perm() != 0o600 {
		t.Fatalf("file mode = %o", fileInfo.Mode().Perm())
	}
	directoryInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if directoryInfo.Mode().Perm() != 0o700 {
		t.Fatalf("directory mode = %o", directoryInfo.Mode().Perm())
	}

	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(); !errors.Is(err, telemetryerrors.ErrInsecureInstallation) {
		t.Fatalf("accepted public file: %v", err)
	}
}

func TestDefaultInstallationPath(t *testing.T) {
	root := t.TempDir()
	switch runtime.GOOS {
	case "darwin":
		t.Setenv("HOME", root)
	case "windows":
		t.Setenv("AppData", root)
	default:
		t.Setenv("XDG_CONFIG_HOME", root)
	}

	path, err := DefaultInstallationPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "installation.json" || filepath.Base(filepath.Dir(path)) != "e2engine" {
		t.Fatalf("DefaultInstallationPath() = %q", path)
	}
}
