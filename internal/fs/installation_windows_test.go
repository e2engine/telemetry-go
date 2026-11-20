//go:build windows

package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	telemetryerrors "github.com/e2engine/telemetry-go/pkg/errors"
	"golang.org/x/sys/windows"
)

func TestWindowsPrivateACL(t *testing.T) {
	path := filepath.Join(privateTestDir(t), "installation.json")
	store := NewInstallationStore(path)
	if err := store.Save(testInstallation); err != nil {
		t.Fatal(err)
	}
	descriptor, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;WD)")
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := descriptor.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(); !errors.Is(err, telemetryerrors.ErrInsecureInstallation) {
		t.Fatalf("accepted public ACL: %v", err)
	}
}

func TestWindowsReplacementWhileReading(t *testing.T) {
	path := filepath.Join(privateTestDir(t), "installation.json")
	store := NewInstallationStore(path)
	if err := store.Save(testInstallation); err != nil {
		t.Fatal(err)
	}
	reader, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if err := store.Save(testInstallation); err != nil {
		t.Fatalf("replacement with active reader: %v", err)
	}
	if got, err := store.Load(); err != nil || got != testInstallation {
		t.Fatalf("replacement = (%v, %v)", got, err)
	}
}
