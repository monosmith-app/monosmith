package pathrefs

import (
	"path"
	"strings"
)

// Relative is the path from directory from to directory to, both
// workspace-relative with forward slashes ("" is the workspace root). The
// result has no leading "./": "../core", "libs/core", or "." for the same
// directory.
func Relative(from, to string) string {
	f, t := split(from), split(to)
	common := 0
	for common < len(f) && common < len(t) && f[common] == t[common] {
		common++
	}
	parts := make([]string, 0, len(f)-common+len(t)-common)
	for range f[common:] {
		parts = append(parts, "..")
	}
	parts = append(parts, t[common:]...)
	if len(parts) == 0 {
		return "."
	}
	return strings.Join(parts, "/")
}

func split(dir string) []string {
	dir = path.Clean(strings.Trim(dir, "/"))
	if dir == "." || dir == "" {
		return nil
	}
	return strings.Split(dir, "/")
}

// Retarget rewrites the references text makes, from a file in directory
// fileDir, to the directory oldRoot so that they name newRoot: "../core"
// becomes "../../packages/core", "./libs/core" becomes "./packages/core". It
// returns the new text and how many references changed.
func Retarget(text, fileDir, oldRoot, newRoot string) (string, int) {
	oldRel, newRel := Relative(fileDir, oldRoot), Relative(fileDir, newRoot)
	if oldRel == "." || oldRel == newRel {
		return text, 0
	}
	total := 0
	if !strings.HasPrefix(oldRel, ".") || strings.HasPrefix(oldRel, "./") {
		// A path below the file's own directory is written with or without "./".
		var n int
		text, n = ReplaceReference(text, "./"+oldRel, "./"+newRel)
		total += n
	}
	var n int
	text, n = ReplaceReference(text, oldRel, newRel)
	total += n
	return text, total
}

// Follow rewrites a relative reference written from directory fileDir (a file
// that does not move) when it names something inside oldRoot, which moves to
// newRoot: "../core/x.cmake" becomes "../../packages/core/x.cmake". Anything
// else (another place, "serde", an absolute path) is returned unchanged.
func Follow(ref, fileDir, oldRoot, newRoot string) string {
	if ref == "" || strings.HasPrefix(ref, "/") || strings.Contains(ref, ":") {
		return ref
	}
	target := path.Join(fileDir, ref)
	if target != oldRoot && !strings.HasPrefix(target, oldRoot+"/") {
		return ref
	}
	rel := Relative(fileDir, path.Join(newRoot, strings.TrimPrefix(target, oldRoot)))
	if strings.HasPrefix(ref, "./") && !strings.HasPrefix(rel, "../") {
		rel = "./" + strings.TrimPrefix(rel, "./")
	}
	if strings.HasSuffix(ref, "/") && !strings.HasSuffix(rel, "/") {
		rel += "/"
	}
	return rel
}

// Rebase returns a reference written relative to directory oldDir so that it
// names the same place from directory newDir. A reference to something inside
// oldDir itself is kept, because that moves along. Anything that is not a
// relative path ("serde", "https://...", an absolute path) is returned
// unchanged.
func Rebase(ref, oldDir, newDir string) string {
	return RebaseFrom(ref, oldDir, oldDir, newDir)
}

// RebaseFrom is Rebase for a file that sits anywhere inside the directory that
// moves: fileDir is the directory the reference is written from, oldRoot the
// directory that moves and newRoot where it goes. A reference to something
// inside oldRoot is kept, because that moves along with the file.
func RebaseFrom(ref, fileDir, oldRoot, newRoot string) string {
	if !strings.HasPrefix(ref, "./") && !strings.HasPrefix(ref, "../") {
		return ref
	}
	target := path.Join(fileDir, ref)
	if oldRoot != "" && (target == oldRoot || strings.HasPrefix(target, oldRoot+"/")) {
		return ref
	}
	if strings.HasPrefix(target, "../") || target == ".." {
		return ref // leaves the workspace
	}
	newFileDir := path.Join(newRoot, strings.TrimPrefix(strings.TrimPrefix(fileDir, oldRoot), "/"))
	rel := Relative(newFileDir, target)
	if strings.HasSuffix(ref, "/") && !strings.HasSuffix(rel, "/") {
		rel += "/"
	}
	if strings.HasPrefix(ref, "./") && !strings.HasPrefix(rel, "../") {
		rel = "./" + strings.TrimPrefix(rel, "./")
	}
	return rel
}
