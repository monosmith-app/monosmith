package architecture

import "github.com/monosmith-app/monosmith/apps/smith/configedit"

// Current is the architecture the workspace at root uses: the first one whose
// table smith.toml has, or None.
func Current(root string) (Architecture, error) {
	for _, a := range catalog {
		if a.Table == "" {
			continue
		}
		on, err := configedit.HasTable(root, a.Table)
		if err != nil {
			return Architecture{}, err
		}
		if on {
			return a, nil
		}
	}
	return Find(None)
}
