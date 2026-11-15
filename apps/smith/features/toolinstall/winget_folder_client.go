package toolinstall

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// wingetSearchDepth is how many folders deep under a package folder the
// program is looked for: dart-sdk\bin\dart.exe is two deep.
const wingetSearchDepth = 3

// placeholder matches a version placeholder of a catalogue package id
// ("PHP.PHP.{major.minor}").
var placeholder = regexp.MustCompile(`\{[^}]*\}`)

// WingetPackagesDir is where winget unpacks a portable package for the user:
// %LOCALAPPDATA%\Microsoft\WinGet\Packages.
//
// Parameters: localAppData is the value of LOCALAPPDATA.
// Returns: the folder, or "" when localAppData is "".
func WingetPackagesDir(localAppData string) string {
	if localAppData == "" {
		return ""
	}
	return filepath.Join(localAppData, "Microsoft", "WinGet", "Packages")
}

// FindWingetFolder looks for a tool's program in the folders where winget
// unpacked a portable package (a zip). winget names such a folder
// <PackageId>_<Source>, for example Google.DartSDK_Microsoft.Winget.Source_8wekyb3d8bbwe.
// It is not on PATH when winget made no link for the program.
//
// Parameters: localAppData is the value of LOCALAPPDATA; tool is the tool, whose
// winget install ways give the package ids (a version placeholder matches any
// version) and whose program and alternatives give the names to look for.
// Returns: the folder that holds <program>.exe, nearest to the package folder
// first, and the newest package first; "" when there is none.
func FindWingetFolder(localAppData string, tool Need) string {
	root := WingetPackagesDir(localAppData)
	if root == "" {
		return ""
	}
	programs := append([]string{tool.Program}, tool.Alternatives...)
	for _, in := range tool.Install {
		if in.Manager != "winget" || in.Package == "" {
			continue
		}
		pattern := placeholder.ReplaceAllString(in.Package, "*")
		matches, _ := filepath.Glob(filepath.Join(root, pattern+"_*"))
		slices.Sort(matches)
		slices.Reverse(matches)
		for _, pkg := range matches {
			if dir := findProgram(pkg, programs); dir != "" {
				return dir
			}
		}
	}
	return ""
}

// findProgram walks a package folder breadth first, at most wingetSearchDepth
// folders deep, and returns the first folder that holds one of the programs.
func findProgram(pkg string, programs []string) string {
	level := []string{pkg}
	for depth := 0; depth <= wingetSearchDepth && len(level) > 0; depth++ {
		var next []string
		for _, dir := range level {
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if e.IsDir() {
					next = append(next, filepath.Join(dir, e.Name()))
					continue
				}
				for _, program := range programs {
					if program != "" && strings.EqualFold(e.Name(), program+".exe") {
						return dir
					}
				}
			}
		}
		level = next
	}
	return ""
}
