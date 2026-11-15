package recipe

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/pelletier/go-toml/v2"
)

// LockFile is the workspace file that pins the recipes loaded from URLs.
const LockFile = "smith.recipes.lock"

// CacheDir is where URL recipes are kept, under the workspace root. It is
// build output: ignored by git, rebuilt from the URLs and checked against the
// lockfile.
var CacheDir = filepath.Join(".smith", "recipes")

const lockHeader = `# Written by smith. It pins the content of every recipe loaded from a URL, so
# a build does not change because a URL's content did. Review a change, then
# accept it with smith upgrade.
`

// LockedRecipe is one pinned URL.
type LockedRecipe struct {
	URL    string `toml:"url"`
	SHA256 string `toml:"sha256"`
}

// Lock is the lockfile's content.
type Lock struct {
	Recipes []LockedRecipe `toml:"recipe"`
}

// find returns the pin for a URL.
func (l Lock) find(url string) (LockedRecipe, bool) {
	for _, r := range l.Recipes {
		if r.URL == url {
			return r, true
		}
	}
	return LockedRecipe{}, false
}

// with returns the lock with a URL pinned to a hash, keeping the entries
// sorted so the file does not change order.
func (l Lock) with(url, sum string) Lock {
	out := Lock{}
	for _, r := range l.Recipes {
		if r.URL != url {
			out.Recipes = append(out.Recipes, r)
		}
	}
	out.Recipes = append(out.Recipes, LockedRecipe{URL: url, SHA256: sum})
	sort.Slice(out.Recipes, func(i, j int) bool { return out.Recipes[i].URL < out.Recipes[j].URL })
	return out
}

// readLock reads the workspace's lockfile; a missing file is an empty lock.
func readLock(root string) (Lock, error) {
	raw, err := os.ReadFile(filepath.Join(root, LockFile))
	if errors.Is(err, fs.ErrNotExist) {
		return Lock{}, nil
	}
	if err != nil {
		return Lock{}, err
	}
	var l Lock
	if err := toml.Unmarshal(raw, &l); err != nil {
		return Lock{}, &ParseError{Source: LockFile, Message: LockFile + ": " + err.Error()}
	}
	return l, nil
}

func writeLock(root string, l Lock) error {
	raw, err := toml.Marshal(l)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, LockFile), append([]byte(lockHeader+"\n"), raw...), 0o644)
}

// sum is the hash a lock records.
func sum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// cachePath is where a URL's content is kept.
func cachePath(root, url string) string {
	return filepath.Join(root, CacheDir, sum([]byte(url))[:16]+".toml")
}

// readCache returns the cached content of a URL, or nil.
func readCache(root, url string) []byte {
	raw, err := os.ReadFile(cachePath(root, url))
	if err != nil {
		return nil
	}
	return raw
}

func writeCache(root, url string, data []byte) error {
	path := cachePath(root, url)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
