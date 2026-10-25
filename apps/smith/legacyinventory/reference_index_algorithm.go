package legacyinventory

import (
	"encoding/json"
	"path"
	"regexp"
	"strings"
)

// maxIndexedSize is the largest file the index reads.
const maxIndexedSize = 2 << 20

var tokenPattern = regexp.MustCompile(`[@A-Za-z0-9_.][@A-Za-z0-9_./-]*`)

// textExtensions are the files that can import or run a package.
var textExtensions = map[string]bool{
	".js": true, ".cjs": true, ".mjs": true, ".jsx": true, ".ts": true, ".cts": true, ".mts": true, ".tsx": true,
	".json": true, ".jsonc": true, ".json5": true, ".yml": true, ".yaml": true, ".sh": true, ".cmd": true, ".ps1": true,
	".toml": true, ".html": true, ".vue": true, ".svelte": true, ".astro": true,
}

// commandExtensions are the files whose words are commands that run tools.
var commandExtensions = map[string]bool{".yml": true, ".yaml": true, ".sh": true, ".cmd": true, ".ps1": true}

// configWords follow the tool's name in the name of its config file.
var configWords = map[string]bool{"config": true, "setup": true, "preset": true, "workspace": true, "resolver": true, "base": true}

var skippedFolders = []string{"node_modules", "dist", "coverage", ".smith", ".nx", ".git", "out-tsc"}

// dependencySections are package.json fields that only list names: a package
// named there is declared, not used.
var dependencySections = []string{"dependencies", "devDependencies", "peerDependencies", "optionalDependencies", "overrides", "resolutions", "peerDependenciesMeta"}

// Index records which words appear in which files that stay, so a question
// like "does anything still run rollup" is a lookup.
type Index struct {
	code     map[string][]string
	commands map[string][]string
}

// BuildIndex reads every file of the tree that can import or run a package.
func BuildIndex(t *Tree) *Index {
	ix := &Index{code: map[string][]string{}, commands: map[string][]string{}}
	for _, f := range t.Files() {
		if !indexable(f) {
			continue
		}
		text := t.Read(f)
		if len(text) > maxIndexedSize {
			continue
		}
		ext := path.Ext(f)
		commandFile := commandExtensions[ext] || strings.HasPrefix(f, ".husky/")
		if path.Base(f) == "package.json" {
			text = manifestWords(text, ix, f)
		} else {
			text = WithoutCommentLines(text, ext)
		}
		ix.add(ix.code, f, text)
		if commandFile {
			ix.add(ix.commands, f, text)
		}
	}
	return ix
}

func indexable(rel string) bool {
	for _, folder := range skippedFolders {
		if strings.HasPrefix(rel, folder+"/") || strings.Contains(rel, "/"+folder+"/") {
			return false
		}
	}
	base := path.Base(rel)
	if base == "package-lock.json" || strings.HasSuffix(base, ".tsbuildinfo") || strings.HasSuffix(base, ".map") {
		return false
	}
	return textExtensions[path.Ext(rel)] || strings.HasPrefix(rel, ".husky/")
}

// manifestWords is a package.json without the sections that only name
// packages; the words of its scripts also go in the commands index.
func manifestWords(text string, ix *Index, file string) string {
	var doc map[string]json.RawMessage
	if json.Unmarshal([]byte(text), &doc) != nil {
		return text
	}
	for _, section := range dependencySections {
		delete(doc, section)
	}
	if raw, ok := doc["scripts"]; ok {
		var scripts map[string]string
		if json.Unmarshal(raw, &scripts) == nil {
			for _, command := range scripts {
				ix.add(ix.commands, file, command)
			}
		}
	}
	out, err := json.Marshal(doc)
	if err != nil {
		return text
	}
	return string(out)
}

func (ix *Index) add(into map[string][]string, file, text string) {
	seen := map[string]bool{}
	note := func(token string) {
		if token == "" || seen[token] {
			return
		}
		seen[token] = true
		into[token] = append(into[token], file)
	}
	for _, token := range tokenPattern.FindAllString(text, -1) {
		token = strings.TrimRight(token, "./-")
		for strings.HasPrefix(token, "./") || strings.HasPrefix(token, "../") {
			token = token[strings.Index(token, "/")+1:]
		}
		note(token)
		// A package named by a deeper path: @scope/name/sub, name/sub.
		parts := strings.Split(token, "/")
		for i := 1; i < len(parts); i++ {
			note(strings.Join(parts[:i], "/"))
		}
		// Words of a dotted name: jest.fn is a use of jest.
		// A config file's name (jest.config.ts) is not.
		if words := strings.Split(token, "."); len(words) > 1 && !strings.Contains(token, "/") && !configWords[words[1]] {
			for _, word := range words {
				note(word)
			}
		}
	}
}

// Files lists up to limit files that mention any of the words.
func (ix *Index) Files(limit int, words ...string) []string {
	var out []string
	seen := map[string]bool{}
	for _, word := range words {
		for _, source := range []map[string][]string{ix.code, ix.commands} {
			for _, f := range source[word] {
				if !seen[f] && len(out) < limit {
					seen[f] = true
					out = append(out, f)
				}
			}
		}
	}
	return out
}

// Mentions reports whether any file mentions any of the words.
func (ix *Index) Mentions(words ...string) bool { return len(ix.Files(1, words...)) > 0 }

// Mentioning lists up to limit files of the tree that mention any of the
// words (comments on a line of their own do not count).
func Mentioning(t *Tree, limit int, words ...string) []string {
	return BuildIndex(t).Files(limit, words...)
}

// Runs reports whether a command file or script mentions any of the words.
func (ix *Index) Runs(words ...string) bool {
	for _, w := range words {
		if len(ix.commands[w]) > 0 {
			return true
		}
	}
	return false
}
