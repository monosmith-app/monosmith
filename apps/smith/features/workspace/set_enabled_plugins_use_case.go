package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	pluginsHeader = regexp.MustCompile(`(?m)^\[plugins\][ \t]*(#.*)?\r?$`)
	enabledLine   = regexp.MustCompile(`(?m)^enabled[ \t]*=[ \t]*\[[^\]]*\][ \t]*(#.*)?\r?$`)
	nextTable     = regexp.MustCompile(`(?m)^\[`)
)

// SetEnabledPlugins writes `enabled = [...]` under [plugins] in smith.toml,
// changing only that line (or adding it, or the table), so comments and every
// other setting survive.
func SetEnabledPlugins(root string, names []string) error {
	path := filepath.Join(root, FileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = fmt.Sprintf("%q", n)
	}
	line := "enabled = [" + strings.Join(quoted, ", ") + "]"
	src := string(raw)
	out, err := setInTable(src, line)
	if err != nil {
		return err
	}
	if _, err := Parse([]byte(out)); err != nil {
		return fmt.Errorf("refusing to write an invalid %s: %w", FileName, err)
	}
	return os.WriteFile(path, []byte(out), 0o644)
}

func setInTable(src, line string) (string, error) {
	header := pluginsHeader.FindStringIndex(src)
	if header == nil {
		if !strings.HasSuffix(src, "\n") {
			src += "\n"
		}
		return src + "\n[plugins]\n" + line + "\n", nil
	}
	bodyStart := header[1]
	bodyEnd := len(src)
	if next := nextTable.FindStringIndex(src[bodyStart:]); next != nil {
		bodyEnd = bodyStart + next[0]
	}
	body := src[bodyStart:bodyEnd]
	if loc := enabledLine.FindStringIndex(body); loc != nil {
		return src[:bodyStart] + body[:loc[0]] + line + body[loc[1]:] + src[bodyEnd:], nil
	}
	return src[:bodyStart] + "\n" + line + src[bodyStart:], nil
}
