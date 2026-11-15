package recipe

import (
	"fmt"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
)

// maxChain bounds how many recipes may extend one another; a longer chain is
// almost certainly a mistake.
const maxChain = 16

// Loader reads recipes from paths and https URLs.
type Loader struct {
	// Root is the workspace root, where URL recipes are cached and pinned. When
	// empty, URL recipes are fetched every time and nothing is written.
	Root string
	// Client fetches URLs; nil uses a default with a timeout.
	Client *http.Client
}

// Loaded is a recipe and its parents.
type Loaded struct {
	// Chain is the recipes involved, the oldest ancestor first and the one
	// that was asked for last.
	Chain []Document
}

// Change is what accepting a URL recipe's current content did.
type Change struct {
	URL string
	// Old is the pinned hash, empty when the URL was not pinned.
	Old string
	New string
}

// Changed reports whether the content differs from what was pinned.
func (c Change) Changed() bool { return c.Old != c.New }

// Load reads a recipe and the chain of recipes it extends. A path in extends
// is relative to the file that has it; a recipe loaded from a URL can extend
// only other URLs. A cycle, a missing parent and a chain of more than 16 are
// errors that name the files.
func (l *Loader) Load(ref string) (Loaded, error) {
	var docs []Document
	seen := map[string]bool{}
	var trail []string
	next, base, from := ref, Source{}, ""
	for {
		if len(docs) >= maxChain {
			return Loaded{}, fmt.Errorf("more than %d recipes extend one another: %s", maxChain, strings.Join(trail, " -> "))
		}
		src, data, err := l.read(next, base)
		if err != nil {
			if from != "" {
				return Loaded{}, fmt.Errorf("%s: %w", from, err)
			}
			return Loaded{}, err
		}
		key := identity(src)
		trail = append(trail, src.Ref)
		if seen[key] {
			return Loaded{}, fmt.Errorf("recipes extend each other in a cycle: %s", strings.Join(trail, " -> "))
		}
		seen[key] = true
		doc, err := ParseRecipe(data, src)
		if err != nil {
			return Loaded{}, err
		}
		docs = append(docs, doc)
		if doc.Recipe.Extends == "" {
			break
		}
		next, base = doc.Recipe.Extends, src
		from = fmt.Sprintf("%s:%d: extends %q", src.Ref, doc.Line("extends"), doc.Recipe.Extends)
	}
	for i, j := 0, len(docs)-1; i < j; i, j = i+1, j-1 {
		docs[i], docs[j] = docs[j], docs[i]
	}
	return Loaded{Chain: docs}, nil
}

// identity names a recipe for cycle detection: a URL, or an absolute path.
func identity(s Source) string {
	if s.URL {
		return s.Ref
	}
	if abs, err := filepath.Abs(s.Ref); err == nil {
		return abs
	}
	return s.Ref
}

// read finds a recipe's source and content: ref as given, or relative to
// the recipe that extends it.
func (l *Loader) read(ref string, base Source) (Source, []byte, error) {
	switch {
	case strings.HasPrefix(ref, "http://"):
		return Source{}, nil, fmt.Errorf("recipes load over https only: %s", ref)
	case strings.HasPrefix(ref, "https://"):
		data, err := l.urlContent(ref, false)
		return Source{Ref: ref, URL: true}, data, err
	case strings.Contains(ref, "://"):
		return Source{}, nil, fmt.Errorf("a recipe is a path or an https URL, not %s", ref)
	case base.URL:
		resolved, err := resolveURL(base.Ref, ref)
		if err != nil {
			return Source{}, nil, err
		}
		data, err := l.urlContent(resolved, false)
		return Source{Ref: resolved, URL: true}, data, err
	}
	path := ref
	if base.Ref != "" && !filepath.IsAbs(ref) {
		path = filepath.Join(filepath.Dir(base.Ref), ref)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Source{}, nil, fmt.Errorf("reading the recipe %s: %w", path, err)
	}
	return Source{Ref: path}, data, nil
}

func resolveURL(base, ref string) (string, error) {
	b, err := neturl.Parse(base)
	if err != nil {
		return "", err
	}
	r, err := neturl.Parse(ref)
	if err != nil {
		return "", err
	}
	return b.ResolveReference(r).String(), nil
}

// urlContent returns a URL recipe's content. When the workspace has pinned
// it and the pinned content is cached, that is used without the network.
// Otherwise it is fetched and, if the lockfile has a hash for it, must match:
// a URL whose content changed is refused until Accept takes the new content.
// A URL not yet pinned is pinned on first use. force fetches whatever is
// pinned and pins what it gets.
func (l *Loader) urlContent(url string, force bool) ([]byte, error) {
	if l.Root == "" {
		return fetch(l.Client, url)
	}
	lock, err := readLock(l.Root)
	if err != nil {
		return nil, err
	}
	pinned, isPinned := lock.find(url)
	if !force && isPinned {
		if cached := readCache(l.Root, url); cached != nil && sum(cached) == pinned.SHA256 {
			return cached, nil
		}
	}
	data, err := fetch(l.Client, url)
	if err != nil {
		return nil, err
	}
	got := sum(data)
	if !force && isPinned && pinned.SHA256 != got {
		return nil, &ChangedError{URL: url, Locked: pinned.SHA256, Got: got}
	}
	if err := writeCache(l.Root, url, data); err != nil {
		return nil, err
	}
	if !isPinned || pinned.SHA256 != got {
		if err := writeLock(l.Root, lock.with(url, got)); err != nil {
			return nil, err
		}
	}
	return data, nil
}

// Accept fetches a URL recipe again and pins what it finds, replacing the
// pinned hash and the cached copy. It is what `smith upgrade` calls to accept
// a recipe that changed (a ChangedError says so); Change says what was pinned
// and what is now.
func (l *Loader) Accept(url string) (Change, error) {
	if l.Root == "" {
		return Change{}, fmt.Errorf("recipes are pinned in a workspace; there is none here")
	}
	lock, err := readLock(l.Root)
	if err != nil {
		return Change{}, err
	}
	old, _ := lock.find(url)
	data, err := l.urlContent(url, true)
	if err != nil {
		return Change{}, err
	}
	return Change{URL: url, Old: old.SHA256, New: sum(data)}, nil
}

// AcceptAll accepts every pinned URL recipe.
func (l *Loader) AcceptAll() ([]Change, error) {
	if l.Root == "" {
		return nil, nil
	}
	lock, err := readLock(l.Root)
	if err != nil {
		return nil, err
	}
	var out []Change
	for _, pinned := range lock.Recipes {
		c, err := l.Accept(pinned.URL)
		if err != nil {
			return out, err
		}
		out = append(out, c)
	}
	return out, nil
}
