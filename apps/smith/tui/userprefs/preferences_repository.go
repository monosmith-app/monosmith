package userprefs

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// Path is the preferences file in the smith home; "" when there is no home.
func Path(home string) string {
	if home == "" {
		return ""
	}
	return filepath.Join(home, FileName)
}

// Load reads the preferences in the smith home. A missing file is no
// preferences at all; a file that cannot be read gives the zero value and the
// reason, so a caller can use the defaults and still say why.
func Load(home string) (Preferences, error) {
	path := Path(home)
	if path == "" {
		return Preferences{}, nil
	}
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Preferences{}, nil
	}
	if err != nil {
		return Preferences{}, err
	}
	var p Preferences
	if err := json.Unmarshal(raw, &p); err != nil {
		return Preferences{}, errors.New(words.Say("userprefs.invalid", "path", path, "reason", err))
	}
	return p, nil
}

// Save writes the preferences, creating the smith home when it is not there
// yet. The file is replaced whole, so a reader never sees half of it.
func Save(home string, p Preferences) error {
	path := Path(home)
	if path == "" {
		return errors.New(words.Say("userprefs.nohome"))
	}
	if err := os.MkdirAll(home, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(home, FileName+".*.tmp")
	if err != nil {
		return err
	}
	_, werr := tmp.Write(append(raw, '\n'))
	cerr := tmp.Close()
	if err := errors.Join(werr, cerr); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return nil
}
