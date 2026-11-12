package toolrun

import (
	"path/filepath"
	"slices"
	"strings"
)

// wingetAlreadyInstalled are the winget return codes that mean the package is
// already on the machine, so there is nothing to install. The list is from
// winget's documented return codes
// (https://github.com/microsoft/winget-cli/blob/master/doc/windows/package-manager/winget/returnCodes.md).
var wingetAlreadyInstalled = []uint32{
	0x8A15002B, // APPINSTALLER_CLI_ERROR_UPDATE_NOT_APPLICABLE: no applicable update
	0x8A150061, // APPINSTALLER_CLI_ERROR_PACKAGE_ALREADY_INSTALLED
	0x8A15010D, // APPINSTALLER_CLI_ERROR_INSTALL_ALREADY_INSTALLED: another version is installed
	0x8A15010E, // APPINSTALLER_CLI_ERROR_INSTALL_DOWNGRADE: a newer version is installed
	0x8A15004F, // APPINSTALLER_CLI_ERROR_UPGRADE_VERSION_NOT_NEWER
}

// WingetAlreadyInstalled reports whether a command is winget and its exit code
// says that the package is already installed (or a newer one is), which is a
// success for an install step and not a failure.
//
// Parameters: argv is the command that ran; code is its exit code as Go
// reports it (on Windows the unsigned 32-bit code).
// Returns: true for winget with one of the documented "already installed" codes.
func WingetAlreadyInstalled(argv []string, code int) bool {
	if len(argv) == 0 || code == 0 {
		return false
	}
	name := strings.ToLower(filepath.Base(argv[0]))
	if name != "winget" && name != "winget.exe" {
		return false
	}
	return slices.Contains(wingetAlreadyInstalled, uint32(code))
}
