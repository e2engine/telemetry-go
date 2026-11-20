package fs

import (
	"errors"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"

	telemetryerrors "github.com/e2engine/telemetry-go/pkg/errors"
)

func privateDACL() (*windows.ACL, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, err
	}
	sddl := "D:P(A;OICI;FA;;;" + user.User.Sid.String() + ")(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)"
	descriptor, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return nil, err
	}
	acl, _, err := descriptor.DACL()
	return acl, err
}

func prepareDirectory(path string) error {
	_, err := os.Lstat(path)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	if mkdirErr := os.MkdirAll(path, 0o700); mkdirErr != nil {
		return mkdirErr
	}
	acl, err := privateDACL()
	if err != nil {
		return err
	}
	return windows.SetNamedSecurityInfo(
		path, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, acl, nil,
	)
}

func makePrivate(file *os.File) error {
	acl, err := privateDACL()
	if err != nil {
		return err
	}
	return windows.SetNamedSecurityInfo(
		file.Name(), windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, acl, nil,
	)
}

func checkPrivate(file *os.File) error {
	descriptor, err := windows.GetSecurityInfo(
		windows.Handle(file.Fd()), windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION,
	)
	if err != nil {
		return err
	}
	acl, _, err := descriptor.DACL()
	if err != nil {
		return err
	}
	if acl == nil {
		return telemetryerrors.ErrInsecureInstallation
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
	system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return err
	}
	admins, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return err
	}
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(acl, i, &ace); err != nil {
			return err
		}
		if ace.Header.AceType == windows.ACCESS_DENIED_ACE_TYPE || ace.Header.AceFlags&windows.INHERIT_ONLY_ACE != 0 {
			continue
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return telemetryerrors.ErrInsecureInstallation
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		if !sid.IsValid() || (!sid.Equals(user.User.Sid) && !sid.Equals(system) && !sid.Equals(admins)) {
			return telemetryerrors.ErrInsecureInstallation
		}
	}
	return nil
}

// FileRenameInfoEx with POSIX semantics replaces the name in one native rename
// operation, including when readers have the previous file open. Do not fall
// back to remove/copy on unsupported Windows versions or filesystems.
func replaceInstallation(_ *os.Root, dir, oldName, newName string) (result error) {
	oldPath, err := windows.UTF16PtrFromString(filepath.Join(dir, oldName))
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(
		oldPath, windows.DELETE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0,
	)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, windows.CloseHandle(handle)) }()
	dest, err := filepath.Abs(filepath.Join(dir, newName))
	if err != nil {
		return err
	}
	name, err := windows.UTF16FromString(dest)
	if err != nil {
		return err
	}
	name = name[:len(name)-1]
	type renameInfo struct {
		Flags          uint32
		RootDirectory  windows.Handle
		FileNameLength uint32
		FileName       [1]uint16
	}
	offset := int(unsafe.Offsetof(renameInfo{}.FileName))
	size := offset + len(name)*2
	// uint64 backing storage preserves native structure alignment on all targets.
	backing := make([]uint64, (size+7)/8)
	info := (*renameInfo)(unsafe.Pointer(&backing[0]))
	info.Flags = windows.FILE_RENAME_REPLACE_IF_EXISTS | windows.FILE_RENAME_POSIX_SEMANTICS
	info.FileNameLength = uint32(len(name) * 2)
	copy(unsafe.Slice(&info.FileName[0], len(name)), name)
	return windows.SetFileInformationByHandle(
		handle, windows.FileRenameInfoEx, (*byte)(unsafe.Pointer(info)), uint32(size),
	)
}
