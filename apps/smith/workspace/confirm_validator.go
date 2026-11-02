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

// validateUpdateCheck checks [workspace] update_check: "on", "off" or empty
// (not chosen yet).
func validateUpdateCheck(cfg *Config) error {
	switch cfg.Workspace.UpdateCheck {
	case "", "on", "off":
		return nil
	}
	return fmt.Errorf("%s: [workspace] update_check is %q; use \"on\" (ask npm once a day for a newer smith) or \"off\" (never ask)", FileName, cfg.Workspace.UpdateCheck)
}