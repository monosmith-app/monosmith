package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// manifestFile is the file that makes a directory a widget project.
const manifestFile = "widget.json"

// manifest is the content of a widget.json.
type manifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Kind        string   `json:"kind"` // app, lib or package; default lib
	Description string   `json:"description"`
	License     string   `json:"license"`
	Uses        []string `json:"uses"` // names of other widgets this one imports
	Stack       string   `json:"stack"`
}

var semver = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

func readManifest(path string) (manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, err
	}
	var m manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return manifest{}, fmt.Errorf("%s: %w", filepath.ToSlash(path), err)
	}
	return m, nil
}

// validate is what the test task checks.
func (m manifest) validate() error {
	if m.Name == "" {
		return fmt.Errorf("%s has no name", manifestFile)
	}
	if !semver.MatchString(m.Version) {
		return fmt.Errorf("%s: version %q is not x.y.z", manifestFile, m.Version)
	}
	return nil
}

// setField rewrites one string field of a widget.json, keeping the order and
// layout of every other key.
func setField(path, key, value string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", filepath.ToSlash(path), err)
	}
	doc.SetString(key, value)
	return os.WriteFile(path, doc.Bytes(), 0o644)
}
