package workspace

import (
	"fmt"
	"strings"
)

// validateTagFormat checks [release] tag_format: it must name the project
// with {name}, or every project would share one prefix; {scope} needs a
// [workspace] scope; and no other placeholder is known.
func validateTagFormat(cfg *Config) error {
	format := cfg.Release.TagFormat
	if format == "" {
		return nil
	}
	if !strings.Contains(format, "{name}") {
		return fmt.Errorf("%s: [release] tag_format %q has no {name}, so every project would share one tag prefix", FileName, format)
	}
	if strings.Contains(format, "{scope}") && cfg.Workspace.Scope == "" {
		return fmt.Errorf("%s: [release] tag_format %q uses {scope}, but [workspace] scope is not set", FileName, format)
	}
	rest := strings.NewReplacer("{scope}", "", "{name}", "").Replace(format)
	if strings.ContainsAny(rest, "{}") {
		return fmt.Errorf("%s: [release] tag_format %q has a placeholder other than {name} and {scope}", FileName, format)
	}
	return nil
}
