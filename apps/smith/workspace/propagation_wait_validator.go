package workspace

import (
	"fmt"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/npmpropagation"
)

// validatePropagationWait refuses a propagation_wait that is not a duration,
// in [release] or in a [projects.<name>.release] table.
func validatePropagationWait(cfg *Config) error {
	if _, err := npmpropagation.ParseWait(cfg.Release.PropagationWait); err != nil {
		return fmt.Errorf("%s: [release] %w", FileName, err)
	}
	names := make([]string, 0, len(cfg.Projects))
	for name := range cfg.Projects {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		r := cfg.Projects[name].Release
		if r == nil {
			continue
		}
		if _, err := npmpropagation.ParseWait(r.PropagationWait); err != nil {
			return fmt.Errorf("%s: [projects.%s.release] %w", FileName, name, err)
		}
	}
	return nil
}
