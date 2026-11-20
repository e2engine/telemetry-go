//go:build !windows

package fs

import (
	"os"

	telemetryerrors "github.com/e2engine/telemetry-go/pkg/errors"
)

func prepareDirectory(path string) error { return os.MkdirAll(path, 0o700) }
func checkPrivate(file *os.File) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.Mode().Perm()&0o077 != 0 {
		return telemetryerrors.ErrInsecureInstallation
	}
	return nil
}
func makePrivate(file *os.File) error { return file.Chmod(0o600) }
func replaceInstallation(root *os.Root, _, oldName, newName string) error {
	return root.Rename(oldName, newName)
}
