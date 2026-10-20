//go:build windows

package security

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

func validatePrivateFile(file *os.File, _ os.FileInfo) error {
	securityDescriptor, err := windows.GetSecurityInfo(
		windows.Handle(file.Fd()),
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION,
	)
	if err != nil {
		return fmt.Errorf("reading private file ACL: %w", err)
	}

	owner, _, err := securityDescriptor.Owner()
	if err != nil || owner == nil {
		return fmt.Errorf("private file owner could not be verified")
	}
	currentUser, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil || currentUser == nil || currentUser.User.Sid == nil || !owner.Equals(currentUser.User.Sid) {
		return fmt.Errorf("private file must be owned by the current user")
	}

	dacl, _, err := securityDescriptor.DACL()
	if err != nil || dacl == nil {
		return fmt.Errorf("private file must have a restrictive ACL")
	}

	for index := uint32(0); index < uint32(dacl.AceCount); index++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, index, &ace); err != nil || ace == nil || ace.Header.AceSize < 12 {
			return fmt.Errorf("private file ACL could not be verified")
		}

		switch ace.Header.AceType {
		case windows.ACCESS_DENIED_ACE_TYPE:
			continue
		case windows.ACCESS_ALLOWED_ACE_TYPE:
			// ACCESS_ALLOWED_ACE and ACCESS_DENIED_ACE share this header and SID layout.
			sidOffset := unsafe.Sizeof(ace.Header) + unsafe.Sizeof(ace.Mask)
			sid := (*windows.SID)(unsafe.Add(unsafe.Pointer(ace), sidOffset))
			if !sid.IsValid() || !privateFilePrincipal(sid, owner, currentUser.User.Sid) {
				return fmt.Errorf("private file ACL grants access to another principal")
			}
		default:
			return fmt.Errorf("private file ACL contains an unsupported access entry")
		}
	}

	return nil
}

func privateFilePrincipal(sid, owner, currentUser *windows.SID) bool {
	return sid.Equals(owner) || sid.Equals(currentUser) ||
		sid.IsWellKnown(windows.WinLocalSystemSid) ||
		sid.IsWellKnown(windows.WinBuiltinAdministratorsSid)
}
