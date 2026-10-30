package releasing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-bundle/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WriteVersion records a release's version in the bundle's app.json, changing
// only the top-level "version" string so the rest of the file keeps its
// formatting. An app.json without one is left alone.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	rel := path.Join(params.Project.Root, discovery.AppFile)
	full := filepath.Join(workspaceRoot, filepath.FromSlash(rel))
	src, err := os.ReadFile(full)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	out, changed, err := setVersion(src, params.Version)
	if err != nil {
		return protocol.VersionResult{}, fmt.Errorf("%s: %w", rel, err)
	}
	if !changed {
		return protocol.VersionResult{Files: []string{}}, nil
	}
	if !params.DryRun {
		if err := os.WriteFile(full, out, 0o644); err != nil {
			return protocol.VersionResult{}, err
		}
	}
	return protocol.VersionResult{Files: []string{rel}}, nil
}

// Publish has no registry to push to: a bundle is released by its tag.
func Publish(protocol.PublishParams) (protocol.PublishResult, error) {
	return protocol.PublishResult{Skipped: "released by its tag"}, nil
}

// setVersion replaces the string value of the top-level "version" key.
func setVersion(src []byte, version string) ([]byte, bool, error) {
	dec := json.NewDecoder(bytes.NewReader(src))
	if _, err := dec.Token(); err != nil { // the opening brace
		return nil, false, err
	}
	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			return nil, false, err
		}
		if key == "version" {
			var old string
			if err := dec.Decode(&old); err != nil {
				return src, false, nil // not a string: nothing to write
			}
			if old == version {
				return src, false, nil
			}
			end := int(dec.InputOffset())
			start := bytes.LastIndexByte(src[:end-1], '"') + 1
			out := append(append(append([]byte{}, src[:start]...), version...), src[end-1:]...)
			return out, true, nil
		}
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return nil, false, err
		}
	}
	return src, false, nil
}
