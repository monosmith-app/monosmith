package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ClearEnabledPlugins removes the `enabled = [...]` line under [plugins] in
// smith.toml, so every built-in plugin is on (the default) and a plugin added
// in a later smith is on too. The rest of the file is left as it is, and a file
// without the line is not written.
//
// root is the workspace folder. It returns an error when smith.toml cannot be
// read or written, or when the result would not parse.
func ClearEnabledPlugins(root string) error {
	path := filepath.Join(root, FileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(raw)
	header := pluginsHeader.FindStringIndex(src)
	if header == nil {
		return nil
	}
	bodyStart, bodyEnd := header[1], len(src)
	if next := nextTable.FindStringIndex(src[bodyStart:]); next != nil {
		bodyEnd = bodyStart + next[0]
	}
	body := src[bodyStart:bodyEnd]
	loc := enabledLine.FindStringIndex(body)
	if loc == nil {
		return nil
	}
	end := loc[1]
	if end < len(body) && body[end] == '\n' {
		end++
	}
	start := loc[0]
	out := src[:bodyStart] + body[:start] + body[end:] + src[bodyEnd:]
	if _, err := Parse([]byte(out)); err != nil {
		return fmt.Errorf("refusing to write an invalid %s: %w", FileName, err)
	}
	if strings.TrimSpace(out) == strings.TrimSpace(src) {
		return nil
	}
	return os.WriteFile(path, []byte(out), 0o644)
}
