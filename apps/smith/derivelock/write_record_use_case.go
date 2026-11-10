package derivelock

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const header = "# smith.lock records who owns each file `smith derive` writes, and a hash of what smith\n# last wrote. Commit it. Do not edit it: use `smith derive --eject <file>` to take a file over.\n\n"

// Write saves the record in the workspace root, entries sorted by path. An
// empty record with no file on disk writes nothing, so a workspace that never
// derives stays without a smith.lock. It returns an error when the file cannot
// be written.
func Write(root string, r Record) error {
	path := filepath.Join(root, FileName)
	if len(r.Files) == 0 {
		if _, err := os.Stat(path); err != nil {
			return nil
		}
	}
	r.Schema = SchemaVersion
	files := slices.Clone(r.Files)
	slices.SortFunc(files, func(a, b Entry) int {
		return cmp.Or(strings.Compare(a.Path, b.Path), strings.Compare(a.Renderer, b.Renderer))
	})
	r.Files = files
	body, err := toml.Marshal(r)
	if err != nil {
		return err
	}
	return os.WriteFile(path, append([]byte(header), body...), 0o644)
}
