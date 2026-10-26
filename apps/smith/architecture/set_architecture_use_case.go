package architecture

import "github.com/monosmith-app/monosmith/apps/smith/configedit"

// Set makes name the architecture of the workspace at root: the tables of the
// other architectures are removed and the chosen one's is added, so smith.toml
// changes in those lines only. Choosing the one already in use changes nothing.
func Set(root, name string) error {
	chosen, err := Find(name)
	if err != nil {
		return err
	}
	for _, a := range catalog {
		if a.Table == "" || a.Name == chosen.Name {
			continue
		}
		on, err := configedit.HasTable(root, a.Table)
		if err != nil {
			return err
		}
		if on {
			if err := configedit.RemoveTable(root, a.Table); err != nil {
				return err
			}
		}
	}
	if chosen.Table == "" {
		return nil
	}
	return configedit.AddTable(root, chosen.Table)
}
