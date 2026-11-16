package functionhost

import (
	"io"
	"os"
	"path/filepath"
)

// KeepZipInDrop copies the zip of a function app into
// <workspace>/dist/drop/<app>/, beside the app settings, so that `smith ci pack`
// keeps it in the pipeline's drop artifact. The zip in the build folder stays,
// for the tasks and the cache that read it there.
//
// Parameters: workspaceRoot is the workspace's folder; app names the app in
// dist/drop; zip is the absolute path of the zip to copy.
// Returns: the copy's path, workspace-relative with forward slashes, or an
// error when the zip cannot be read or the copy cannot be written.
func KeepZipInDrop(workspaceRoot, app, zip string) (string, error) {
	rel := filepath.ToSlash(filepath.Join("dist", "drop", app, filepath.Base(zip)))
	target := filepath.Join(workspaceRoot, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	in, err := os.Open(zip)
	if err != nil {
		return "", err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(target)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return "", err
	}
	return rel, out.Close()
}
