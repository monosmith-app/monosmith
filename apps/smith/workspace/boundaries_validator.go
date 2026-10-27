package workspace

import (
	"fmt"
	"strings"
)

// validateBoundaries checks [boundaries]: a rule says who it is about and what
// it allows or denies, and an exception names both projects and why.
func validateBoundaries(cfg *Config) error {
	for i, r := range cfg.Boundaries.Rules {
		at := fmt.Sprintf("%s: [[boundaries.rules]] number %d", FileName, i+1)
		if strings.TrimSpace(r.From) == "" {
			return fmt.Errorf("%s has no from; name the projects the rule is about (a name, a glob, kind:app, lang:go or tag:ui)", at)
		}
		if len(r.Allow) == 0 && len(r.Deny) == 0 {
			return fmt.Errorf("%s (from %q) has neither allow nor deny, so it says nothing", at, r.From)
		}
		for _, selector := range append(append([]string{}, r.Allow...), r.Deny...) {
			if strings.TrimSpace(selector) == "" {
				return fmt.Errorf("%s (from %q) lists an empty project name", at, r.From)
			}
		}
	}
	for i, e := range cfg.Boundaries.Exceptions {
		at := fmt.Sprintf("%s: [[boundaries.exceptions]] number %d", FileName, i+1)
		if strings.TrimSpace(e.From) == "" || strings.TrimSpace(e.To) == "" {
			return fmt.Errorf("%s needs from and to: the project that depends and the one it depends on", at)
		}
		if strings.TrimSpace(e.Reason) == "" {
			return fmt.Errorf("%s (%s -> %s) needs a reason, so the next reader knows why it may stay", at, e.From, e.To)
		}
	}
	return nil
}
