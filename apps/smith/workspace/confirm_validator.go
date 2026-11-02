package workspace

import "fmt"

// validateConfirm checks [workspace] confirm: "always" or "risky", or empty
// (which means always).
func validateConfirm(cfg *Config) error {
	switch cfg.Workspace.Confirm {
	case "", "always", "risky":
		return nil
	}
	return fmt.Errorf("%s: [workspace] confirm is %q; use \"always\" (every plan asks) or \"risky\" (only a plan that deletes or edits files asks)", FileName, cfg.Workspace.Confirm)
}
