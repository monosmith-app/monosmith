package shiplink

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// linkFile and removeFile are os.Link and os.Remove; a test replaces them to
// play a file system without hard links or a program Windows holds open.
var (
	linkFile   = os.Link
	removeFile = os.Remove
)

// Ship puts what a task names in protocol.Task.Ship into the shipped folder of
// the project at projectRoot (protocol.ProjectBuildDir). Each glob is relative
// to the project; each match goes under its base name, a folder file by file.
// A file already linked to the tool's current file stays; a stale one is
// replaced. Ship never fails a build: what it could not do is in the Report.
//
// Parameters: workspaceRoot is the absolute workspace folder; projectRoot the
// workspace-relative project folder; patterns the task's Ship globs.
// It returns what it shipped, kept or could not ship.
func Ship(workspaceRoot, projectRoot string, patterns []string) Report {
	folder := protocol.ProjectBuildDir(projectRoot)
	r := Report{Folder: folder}
	if len(patterns) == 0 {
		return r
	}
	matched := false
	for _, p := range patterns {
		pattern := path.Clean(path.Join(projectRoot, p))
		if !fs.ValidPath(pattern) {
			continue // outside the workspace
		}
		matches, _ := fs.Glob(os.DirFS(workspaceRoot), pattern)
		for _, m := range matches {
			if m == folder || strings.HasPrefix(m, folder+"/") {
				continue // already in the shipped folder
			}
			matched = true
			shipMatch(workspaceRoot, m, path.Join(folder, path.Base(m)), &r)
		}
	}
	r.NothingMatched = !matched
	return r
}

// shipMatch ships one match, a file or every file of a folder, to dst.
func shipMatch(workspaceRoot, src, dst string, r *Report) {
	from := filepath.Join(workspaceRoot, filepath.FromSlash(src))
	info, err := os.Stat(from)
	if err != nil {
		r.Failed = append(r.Failed, Failure{File: dst, Reason: err.Error()})
		return
	}
	if !info.IsDir() {
		shipFile(workspaceRoot, src, dst, info, r)
		return
	}
	_ = filepath.WalkDir(from, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(from, p)
		fi, err := os.Stat(p)
		if err != nil {
			return nil
		}
		shipFile(workspaceRoot, path.Join(src, filepath.ToSlash(rel)), path.Join(dst, filepath.ToSlash(rel)), fi, r)
		return nil
	})
}

// shipFile links src to dst (both workspace-relative), or copies it when the
// link fails. A dst that is already src stays; a stale dst that cannot be
// removed (a running program on Windows) stays too, and is reported.
func shipFile(workspaceRoot, src, dst string, info fs.FileInfo, r *Report) {
	from := filepath.Join(workspaceRoot, filepath.FromSlash(src))
	to := filepath.Join(workspaceRoot, filepath.FromSlash(dst))
	if old, err := os.Lstat(to); err == nil {
		if os.SameFile(info, old) {
			r.Shipped++
			return
		}
		if err := removeFile(to); err != nil && !errors.Is(err, fs.ErrNotExist) {
			r.InUse = append(r.InUse, dst)
			return
		}
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		r.Failed = append(r.Failed, Failure{File: dst, Reason: err.Error()})
		return
	}
	if linkFile(from, to) == nil {
		r.Shipped++
		return
	}
	if err := copyFile(from, to, info.Mode().Perm()); err != nil {
		r.Failed = append(r.Failed, Failure{File: dst, Reason: err.Error()})
		return
	}
	r.Shipped++
	r.Copied++
}

// copyFile writes a copy of from at to, with the mode of the original.
func copyFile(from, to string, mode fs.FileMode) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
