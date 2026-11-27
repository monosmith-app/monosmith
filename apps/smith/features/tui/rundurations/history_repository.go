package rundurations

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// FileName is the history's file in the workspace's local cache folder.
const FileName = "tui-durations.json" // words:literal

// Path is the history's file in the local cache folder dir; "" for no folder.
func Path(dir string) string {
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, FileName)
}

// Load reads the history in the local cache folder dir. A missing file, or one
// that cannot be read, is an empty history: an estimate is never needed.
//
// Parameters: dir is the local cache folder.
// Returns: the history.
func Load(dir string) History {
	path := Path(dir)
	if path == "" {
		return History{}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return History{}
	}
	var h History
	if json.Unmarshal(raw, &h) != nil || h == nil {
		return History{}
	}
	return h
}

// Save writes the history in the local cache folder dir, which it makes when it
// is not there yet. The file is replaced whole.
//
// Parameters: dir is the local cache folder; h is the history.
// Returns: an error when the file cannot be written.
func Save(dir string, h History) error {
	path := Path(dir)
	if path == "" {
		return errors.New(words.Say("rundurations.nodir"))
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	raw, err := json.Marshal(h)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, FileName+".*.tmp")
	if err != nil {
		return err
	}
	_, werr := tmp.Write(raw)
	cerr := tmp.Close()
	if err := errors.Join(werr, cerr); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		if errors.Is(err, fs.ErrExist) {
			return nil
		}
		return err
	}
	return nil
}
