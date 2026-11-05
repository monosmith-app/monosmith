package configedit

import (
	"fmt"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Set changes one setting of the smith.toml at root to args, checked against
// the schema; the whole file must still load before it is written.
func Set(root, path string, args []string) error {
	f, err := Find(path)
	if err != nil {
		return err
	}
	value, err := f.Encode(args)
	if err != nil {
		return err
	}
	return edit(root, func(src string) (string, error) {
		return SetKey(src, f.Table(), f.Key(), value), nil
	})
}

// taskOutputsPath is the path of a task's outputs override in a project's
// table: projects.<name>.tasks.<task>.outputs, the name quoted when it is not a
// bare key. It is the one table setting `config unset` reaches, so a fix can
// hand the build's folder back to the plugin.
var taskOutputsPath = regexp.MustCompile(`^projects\.([A-Za-z0-9_-]+|"[^"]+")\.tasks\.[A-Za-z0-9_:-]+\.outputs$`)

// Unset removes one setting, so it takes its default. Besides the settings
// smith config lists, it removes the outputs a project's task overrides
// (projects.<name>.tasks.<task>.outputs), so the plugin's own outputs apply.
func Unset(root, path string) error {
	if taskOutputsPath.MatchString(path) {
		table := strings.TrimSuffix(path, ".outputs")
		return edit(root, func(src string) (string, error) {
			out, removed := UnsetKey(src, table, "outputs")
			if !removed {
				return "", fmt.Errorf("%s is not set; it already has its default", path)
			}
			return out, nil
		})
	}
	f, err := Find(path)
	if err != nil {
		return err
	}
	return edit(root, func(src string) (string, error) {
		out, removed := UnsetKey(src, f.Table(), f.Key())
		if !removed {
			return "", fmt.Errorf("%s is not set; it already has its default", path)
		}
		return out, nil
	})
}

// Get returns a setting's value as written in smith.toml; set is false when
// it has its default.
func Get(root, path string) (value string, set bool, err error) {
	f, err := Find(path)
	if err != nil {
		return "", false, err
	}
	raw, err := os.ReadFile(filepath.Join(root, workspace.FileName))
	if err != nil {
		return "", false, err
	}
	value, set = GetKey(string(raw), f.Table(), f.Key())
	return value, set, nil
}

func edit(root string, change func(string) (string, error)) error {
	file := filepath.Join(root, workspace.FileName)
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	out, err := change(string(raw))
	if err != nil {
		return err
	}
	if _, err := workspace.Parse([]byte(out)); err != nil {
		return fmt.Errorf("refusing to write an invalid %s: %w", workspace.FileName, err)
	}
	return lineendings.WriteFile(file, []byte(out), 0o644)
}
