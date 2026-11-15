package toolinstall

import "strings"

// userPathScript adds a folder to the Path of the Windows user. It reads the
// raw value (DoNotExpandEnvironmentNames) and writes it back as an expandable
// string, so entries such as %USERPROFILE%\bin keep working. A folder that is
// already there is not added again. Setting and removing a temporary user
// variable makes Windows tell the open programs that the environment changed.
const userPathScript = `$d='%s'; ` +
	`$k=[Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment',$true); ` +
	`$v=[string]$k.GetValue('Path','',[Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames); ` +
	`if(($v -split ';') -notcontains $d){ $k.SetValue('Path',((@($v.TrimEnd(';'),$d) | Where-Object {$_ -ne ''}) -join ';'),[Microsoft.Win32.RegistryValueKind]::ExpandString) }; ` +
	`$k.Close(); ` +
	`[Environment]::SetEnvironmentVariable('SMITH_PATH_CHANGED','1','User'); ` +
	`[Environment]::SetEnvironmentVariable('SMITH_PATH_CHANGED',$null,'User')`

// UserPathScript is the PowerShell text that adds a folder to the Path of the
// Windows user, as a person pastes it into PowerShell.
//
// Parameters: dir is the folder; a single quote in it is doubled, as a
// PowerShell single-quoted string needs.
// Returns: the script.
func UserPathScript(dir string) string {
	return strings.Replace(userPathScript, "%s", strings.ReplaceAll(dir, "'", "''"), 1)
}

// UserPathCommand is the command smith runs to add a folder to the Path of the
// Windows user: PowerShell with UserPathScript. No shell parses it first.
//
// Parameters: dir is the folder.
// Returns: the program and its arguments.
func UserPathCommand(dir string) []string {
	return []string{"powershell", "-NoProfile", "-Command", UserPathScript(dir)}
}
