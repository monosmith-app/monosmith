package templating

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"
)

// ErrExists is returned when a target file already exists.
var ErrExists = errors.New("file already exists")

// Write renders every template in files (target path → template text) under
// root with data, creating directories as needed. Target paths may themselves
// contain template actions. It returns the written paths, forward-slashed and
// relative to root, in sorted order. Nothing is written if any target exists.
func Write(root string, files map[string]string, data any) ([]string, error) {
	type rendered struct {
		path    string
		content []byte
	}
	targets := make([]string, 0, len(files))
	for p := range files {
		targets = append(targets, p)
	}
	sort.Strings(targets)

	var out []rendered
	for _, target := range targets {
		rawName, err := render("path", target, data)
		if err != nil {
			return nil, err
		}
		name := path.Clean(string(rawName))
		body, err := render(name, files[target], data)
		if err != nil {
			return nil, err
		}
		full := filepath.Join(root, filepath.FromSlash(name))
		if _, err := os.Stat(full); err == nil {
			return nil, fmt.Errorf("%s: %w", name, ErrExists)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		out = append(out, rendered{name, body})
	}

	written := make([]string, 0, len(out))
	for _, r := range out {
		full := filepath.Join(root, filepath.FromSlash(r.path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return written, err
		}
		content := []byte(lineendings.Convert(string(r.content), lineendings.NewFileCRLF(root, r.path)))
		if err := os.WriteFile(full, content, 0o644); err != nil {
			return written, err
		}
		written = append(written, r.path)
	}
	return written, nil
}

func render(name, text string, data any) ([]byte, error) {
	tmpl, err := template.New(name).Delims("[[", "]]").Option("missingkey=error").Parse(text)
	if err != nil {
		return nil, fmt.Errorf("template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("template %s: %w", name, err)
	}
	return buf.Bytes(), nil
}
