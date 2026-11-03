package derive

import (
	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// oneOwnerPerFile makes one renderer own each file that renderers of different
// kinds both claim, for example a whole tsconfig.json from a template and some
// keys of it from the shape of the siblings. The renderer that the record lists
// for the file wins. Without a record, the renderer of the whole file wins over
// one that derives keys. Renderers that all derive keys of the same file are
// left alone: each owns its own keys. Disputes between whole files stay for
// PlanFor to report.
func oneOwnerPerFile(rec derivelock.Record, proposals []Proposal) []Proposal {
	byPath := map[string][]Proposal{}
	for _, p := range proposals {
		byPath[p.File.Path] = append(byPath[p.File.Path], p)
	}
	owner := map[string]string{}
	for path, claims := range byPath {
		var whole, recorded []string
		renderers := map[string]bool{}
		for _, c := range claims {
			renderers[c.File.Renderer] = true
			if c.File.Mode == protocol.DeriveFile {
				whole = append(whole, c.File.Renderer)
			}
			if rec.Find(path, c.File.Renderer) != nil {
				recorded = append(recorded, c.File.Renderer)
			}
		}
		if len(renderers) < 2 || len(whole) == 0 {
			continue
		}
		switch {
		case len(recorded) == 1:
			owner[path] = recorded[0]
		case len(recorded) == 0 && len(unique(whole)) == 1:
			owner[path] = whole[0]
		}
	}
	if len(owner) == 0 {
		return proposals
	}
	var out []Proposal
	for _, p := range proposals {
		if who, claimed := owner[p.File.Path]; !claimed || p.File.Renderer == who {
			out = append(out, p)
		}
	}
	return out
}

// unique is the distinct values of s.
func unique(s []string) map[string]bool {
	set := map[string]bool{}
	for _, v := range s {
		set[v] = true
	}
	return set
}
