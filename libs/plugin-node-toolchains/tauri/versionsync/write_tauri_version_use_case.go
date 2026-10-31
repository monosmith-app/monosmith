package versionsync

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WriteVersion records a release's version in the two files of a Tauri app
// that carry one beside package.json (which the node plugin writes):
// src-tauri/Cargo.toml and src-tauri/tauri.conf.json. A tauri.conf.json whose
// version is a path ("../package.json") is left alone, Tauri reads it itself.
// A project without a Tauri app has nothing to write.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	res := protocol.VersionResult{Files: []string{}}
	if !tauriapp.Detect(workspaceRoot, params.Project.Root) {
		return res, nil
	}
	for _, f := range []struct {
		name string
		set  func(string, string) (string, bool)
	}{{"Cargo.toml", setCargoVersion}, {"tauri.conf.json", setConfVersion}} {
		rel := path.Join(params.Project.Root, tauriapp.SrcDir, f.name)
		full := filepath.Join(workspaceRoot, filepath.FromSlash(rel))
		src, err := os.ReadFile(full)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return protocol.VersionResult{}, err
		}
		out, changed := f.set(strings.ReplaceAll(string(src), "\r\n", "\n"), params.Version)
		if !changed {
			continue
		}
		if strings.Contains(string(src), "\r\n") {
			out = strings.ReplaceAll(out, "\n", "\r\n")
		}
		if !params.DryRun {
			if err := os.WriteFile(full, []byte(out), 0o644); err != nil {
				return protocol.VersionResult{}, err
			}
		}
		res.Files = append(res.Files, rel)
	}
	return res, nil
}
