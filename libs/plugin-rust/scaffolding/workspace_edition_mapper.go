package scaffolding

import (
	"regexp"
	"sort"
	"strings"
)

// DefaultEdition is the Rust edition a new workspace shares in its
// [workspace.package] table.
const DefaultEdition = "2021"

var (
	packageHeader  = regexp.MustCompile(`(?m)^\[workspace\.package\][ \t]*(#.*)?\r?\n?`)
	nextTable      = regexp.MustCompile(`(?m)^\[`)
	literalEdition = regexp.MustCompile(`(?m)^edition[ \t]*=[ \t]*"([^"]+)"[ \t]*(#.*)?(\r?\n|$)`)
	editionOfTable = regexp.MustCompile(`(?m)^edition[ \t]*=[ \t]*"([^"]+)"`)
)

// inheritedEdition is the line a crate uses to take the shared edition.
const inheritedEdition = "edition.workspace = true"

// WorkspaceEdition is the edition the root Cargo.toml shares in its
// [workspace.package] table, or "" when it shares none.
func WorkspaceEdition(rootSrc string) string {
	h := packageHeader.FindStringIndex(rootSrc)
	if h == nil {
		return ""
	}
	body := rootSrc[h[1]:]
	if n := nextTable.FindStringIndex(body); n != nil {
		body = body[:n[0]]
	}
	if m := editionOfTable.FindStringSubmatch(body); m != nil {
		return m[1]
	}
	return ""
}

// CrateEdition is the edition a crate's manifest sets for itself, or "" when
// it sets none (it inherits one, or takes cargo's default).
func CrateEdition(crateSrc string) string {
	if m := literalEdition.FindStringSubmatch(crateSrc); m != nil {
		return m[1]
	}
	return ""
}

// ShareEdition moves the edition the crates repeat into the root's
// [workspace.package] table. The shared edition is the root's own when it has
// one, else the one most crates set (the lowest on a tie). A crate with that
// edition inherits it (edition.workspace = true); a crate with another edition
// keeps its own. crates maps a crate directory to its manifest text. It returns
// the new root text, the new text of each crate it changed, and the shared
// edition; the edition is "" when nothing needs to change.
func ShareEdition(rootSrc string, crates map[string]string) (root string, changed map[string]string, edition string) {
	edition = WorkspaceEdition(rootSrc)
	count := map[string]int{}
	for _, src := range crates {
		if e := CrateEdition(src); e != "" {
			count[e]++
		}
	}
	if len(count) == 0 {
		return rootSrc, nil, ""
	}
	if edition == "" {
		var seen []string
		for e := range count {
			seen = append(seen, e)
		}
		sort.Strings(seen)
		edition = seen[0]
		for _, e := range seen {
			if count[e] > count[edition] {
				edition = e
			}
		}
	}
	changed = map[string]string{}
	for dir, src := range crates {
		if CrateEdition(src) == edition {
			changed[dir] = literalEdition.ReplaceAllString(src, inheritedEdition+"\n")
		}
	}
	return withWorkspaceEdition(rootSrc, edition), changed, edition
}

// withWorkspaceEdition makes the root's [workspace.package] table set the
// edition: it adds the table, or the key to a table that has none.
func withWorkspaceEdition(rootSrc, edition string) string {
	if WorkspaceEdition(rootSrc) != "" {
		return rootSrc
	}
	line := "edition = \"" + edition + "\"\n"
	if h := packageHeader.FindStringIndex(rootSrc); h != nil {
		head := rootSrc[:h[1]]
		if !strings.HasSuffix(head, "\n") {
			head += "\n"
		}
		return head + line + rootSrc[h[1]:]
	}
	out := rootSrc
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return out + "\n[workspace.package]\n" + line
}
