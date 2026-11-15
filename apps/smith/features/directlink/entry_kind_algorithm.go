package directlink

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// maxLauncherBytes is the size above which an .exe is a program and not a launcher.
const maxLauncherBytes = 1 << 20

// ours is the text that tells an entry belongs to a smith installation: the
// name of the package in a path (node_modules/monosmith, site-packages/monosmith).
const ours = "monosmith"

// classify says what the file at path is, and whether it belongs to a smith
// installation whose program is exe. An entry that is not smith's, such as another
// program named smith, is not an entry: ok is false.
//
// Parameters: path is a file named smith on PATH; exe is the installed program;
// runDir is the folder of the links an entry already changed calls, and target is the
// program an entry of this version should call: an entry that calls it is Direct.
// Returns: the kind, and true when the entry is smith's.
func classify(path, exe, runDir, target string) (Kind, bool) {
	info, err := os.Lstat(path)
	if err != nil || info.IsDir() {
		return "", false
	}
	if info.Mode()&os.ModeSymlink != 0 {
		real, err := filepath.EvalSymlinks(path)
		if err != nil || (!strings.Contains(real, ours) && !sameFile(real, exe)) {
			return "", false
		}
		if sameFile(real, exe) {
			return Direct, true
		}
		return Symlink, true
	}
	if sameFile(path, exe) {
		return Direct, true
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	head := raw[:min(len(raw), 1<<16)]
	binary := bytes.HasPrefix(raw, []byte("MZ"))
	// An entry smith changed calls the program of its version: Direct. One that calls
	// a link of another version is smith's too, and is changed again.
	if !binary && target != "" && (bytes.Contains(head, []byte(target)) || bytes.Contains(head, []byte(filepath.ToSlash(target)))) {
		return Direct, true
	}
	mine := isOurs(head) || (runDir != "" && (bytes.Contains(head, []byte(runDir)) || bytes.Contains(head, []byte(filepath.ToSlash(runDir)))))
	switch strings.ToLower(filepath.Ext(path)) {
	case ".cmd", ".bat":
		return CmdShim, mine
	case ".ps1":
		return Ps1Shim, mine
	case ".exe":
		// A launcher is small. A large program is a full smith already, and a
		// different install: it is not an entry.
		return Launcher, binary && isOurs(raw) && len(raw) < maxLauncherBytes
	}
	if bytes.HasPrefix(raw, []byte("#!")) {
		return ShShim, mine
	}
	return "", false
}

// isOurs is true when the text names the smith package.
func isOurs(raw []byte) bool { return bytes.Contains(raw, []byte(ours)) }

// sameFile is true when both paths are the same file (a link of it too).
func sameFile(a, b string) bool {
	x, errA := os.Stat(a)
	y, errB := os.Stat(b)
	return errA == nil && errB == nil && os.SameFile(x, y)
}
