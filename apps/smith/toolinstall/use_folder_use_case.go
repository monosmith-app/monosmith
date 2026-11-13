package toolinstall

import (
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// ToolFoldersFile is the file of this user where smith records the folders of
// tools that are not on PATH (toolrun.ToolFoldersFile under smith's home).
func ToolFoldersFile() string { return toolrun.ToolFoldersFile(plugintrust.Home()) }

// UseFolder finds the folder where winget unpacked a tool and records it, so
// every later smith command puts the folder on its own PATH. It changes
// neither the user's nor the machine's PATH.
//
// Parameters: need is the tool; localAppData is the value of LOCALAPPDATA;
// file is ToolFoldersFile.
// Returns: the folder, or an error when no folder holds the program or the
// record cannot be written.
func UseFolder(need Need, localAppData, file string) (string, error) {
	dir := FindWingetFolder(localAppData, need)
	if dir == "" {
		return "", problem.New(words.Say("toolfolder.use.none", "program", need.Program),
			words.Say("toolfolder.use.none.remedy", "tool", need.Name))
	}
	if err := toolrun.RecordToolFolder(file, need.Name, dir); err != nil {
		return "", err
	}
	return dir, toolrun.UseToolFolders(file)
}
