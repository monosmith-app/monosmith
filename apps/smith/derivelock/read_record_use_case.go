package derivelock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Read loads smith.lock from the workspace root. A workspace with no record
// has an empty one. It returns an error when the file is not valid TOML or
// was written by a newer smith.
func Read(root string) (Record, error) {
	raw, err := os.ReadFile(filepath.Join(root, FileName))
	if errors.Is(err, os.ErrNotExist) {
		return Record{Schema: SchemaVersion}, nil
	}
	if err != nil {
		return Record{}, err
	}
	var r Record
	if err := toml.Unmarshal(raw, &r); err != nil {
		return Record{}, fmt.Errorf("%s is not valid TOML: %w", FileName, err)
	}
	if r.Schema > SchemaVersion {
		return Record{}, fmt.Errorf("%s has schema %d, this smith reads up to %d: upgrade smith", FileName, r.Schema, SchemaVersion)
	}
	r.Schema = SchemaVersion
	return r, nil
}
