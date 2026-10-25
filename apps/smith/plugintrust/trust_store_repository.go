package plugintrust

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Store is the file of trust decisions in the user's smith home.
type Store struct {
	// Path is the JSON file; it is created on the first decision.
	Path string
}

type storeFile struct {
	Version int      `json:"version"`
	Trusted []Record `json:"trusted"`
}

// OpenStore is the store in the user's smith home.
func OpenStore() (*Store, error) {
	path := StorePath()
	if path == "" {
		return nil, errors.New("cannot find your home directory to keep trust decisions in; set SMITH_HOME")
	}
	return &Store{Path: path}, nil
}

// Find returns the decision for a plugin in a workspace.
func (s *Store) Find(workspace, plugin string) (Record, bool, error) {
	file, err := s.load()
	if err != nil {
		return Record{}, false, err
	}
	key := workspaceKey(workspace)
	for _, r := range file.Trusted {
		if r.Plugin == plugin && workspaceKey(r.Workspace) == key {
			return r, true, nil
		}
	}
	return Record{}, false, nil
}

// Trust records that the user trusts a plugin as fingerprinted, replacing an
// earlier decision for the same workspace and plugin.
func (s *Store) Trust(workspace, plugin string, fp Fingerprint) error {
	file, err := s.load()
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(workspace)
	if err != nil {
		return err
	}
	file.Trusted = without(file.Trusted, workspace, plugin)
	file.Trusted = append(file.Trusted, Record{Workspace: abs, Plugin: plugin, Fingerprint: fp, TrustedAt: time.Now().UTC().Truncate(time.Second)})
	return s.save(file)
}

// Untrust removes the decision for a plugin in a workspace and reports
// whether there was one.
func (s *Store) Untrust(workspace, plugin string) (bool, error) {
	file, err := s.load()
	if err != nil {
		return false, err
	}
	rest := without(file.Trusted, workspace, plugin)
	if len(rest) == len(file.Trusted) {
		return false, nil
	}
	file.Trusted = rest
	return true, s.save(file)
}

func without(records []Record, workspace, plugin string) []Record {
	key := workspaceKey(workspace)
	var out []Record
	for _, r := range records {
		if r.Plugin == plugin && workspaceKey(r.Workspace) == key {
			continue
		}
		out = append(out, r)
	}
	return out
}

func (s *Store) load() (storeFile, error) {
	raw, err := os.ReadFile(s.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return storeFile{Version: 1}, nil
	}
	if err != nil {
		return storeFile{}, err
	}
	var file storeFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return storeFile{}, fmt.Errorf("%s is not valid JSON (%w); delete it to forget every trust decision", s.Path, err)
	}
	return file, nil
}

func (s *Store) save(file storeFile) error {
	file.Version = 1
	raw, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.Path), ".trust-*.json")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(append(raw, '\n')); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), s.Path); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return nil
}

// workspaceKey is the form two spellings of the same directory share.
func workspaceKey(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if real, err := filepath.EvalSymlinks(path); err == nil {
		path = real
	}
	path = filepath.Clean(path)
	if runtime.GOOS == "windows" {
		path = strings.ToLower(path)
	}
	return path
}
