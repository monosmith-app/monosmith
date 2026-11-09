package siblingfiles

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Renderer is the id of the renderer of files derived from sibling projects.
const Renderer = "node.sibling-shape"

// Version is the version of the renderer.
const Version = 1

// Input is what Derive needs.
type Input struct {
	// Root is the workspace root.
	Root string
	// Projects are the projects the plugin serves.
	Projects []protocol.Project
	// Owned are the paths the record says smith owns.
	Owned []string
	// References gives the tsconfig references of a project.
	References References
}

// Result is what Derive finds.
type Result struct {
	// Files are the files to derive: a whole file, or the keys of a JSON file.
	Files []protocol.DerivedFile
	// Diverging are the files that differ from the shape of their siblings.
	Diverging []protocol.DivergentFile
}

// group is the members that share a project kind and a file name.
type group struct {
	kind    protocol.ProjectKind
	name    string
	members []*member
}

// Derive finds the config files that sibling projects share (ADR 0050) and
// answers with what to derive for each project that has the file, or that smith
// owns and that is gone. A file in the shape of the convention is a whole file.
// A JSON file that has the shape in some top-level keys only is derived key by
// key. Anything else is the project's own and is reported as diverging. It reads
// files under in.Root and writes none.
func Derive(in Input) Result {
	var res Result
	for _, g := range collect(in) {
		derived := g.derive(in.References)
		res.Files = append(res.Files, derived.Files...)
		res.Diverging = append(res.Diverging, derived.Diverging...)
	}
	return res
}

// collect lists the config files of the projects, grouped by kind and name.
func collect(in Input) []*group {
	groups := map[string]*group{}
	add := func(p protocol.Project, name string, raw []byte, exists bool) {
		id := string(p.Kind) + "\x00" + name
		if groups[id] == nil {
			groups[id] = &group{kind: p.Kind, name: name}
		}
		rel := path.Join(p.Root, name)
		groups[id].members = append(groups[id].members, &member{project: p, rel: rel, raw: raw, exists: exists, owned: slices.Contains(in.Owned, rel)})
	}
	for _, p := range in.Projects {
		seen := map[string]bool{}
		entries, _ := os.ReadDir(filepath.Join(in.Root, filepath.FromSlash(p.Root)))
		for _, e := range entries {
			if e.IsDir() || !IsConfigFile(e.Name()) {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(in.Root, filepath.FromSlash(p.Root), e.Name()))
			if err != nil {
				continue
			}
			seen[e.Name()] = true
			add(p, e.Name(), raw, true)
		}
		for _, owned := range in.Owned {
			if name := path.Base(owned); path.Dir(owned) == path.Clean(p.Root) && IsConfigFile(name) && !seen[name] {
				add(p, name, nil, false)
			}
		}
	}
	ids := make([]string, 0, len(groups))
	for id := range groups {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	out := make([]*group, 0, len(ids))
	for _, id := range ids {
		out = append(out, groups[id])
	}
	return out
}

// derive compares the files of the group and answers for each member.
func (g *group) derive(refs References) Result {
	var voters []*member
	for _, m := range g.members {
		m.analyse(refs)
		if m.exists {
			voters = append(voters, m)
		}
	}
	var res Result
	if len(voters) < MinSiblings {
		return res
	}
	whole := wholeShape(voters)
	majority := keyMajorities(voters)
	for _, m := range g.members {
		switch {
		case whole != nil && (!m.exists || m.fits(whole)):
			if content, ok := renderWhole(whole, m, refs); ok {
				res.Files = append(res.Files, protocol.DerivedFile{Project: m.project.Name, Path: m.rel, Mode: protocol.DeriveFile, Renderer: Renderer, Version: Version, Content: content})
			}
		case !m.exists && len(majority) > 0 && whole == nil:
			res.Files = append(res.Files, g.keysFile(m, majority, refs, nil))
		case m.isJSON && len(majority) > 0:
			keys, differing := renderKeys(majority, m, refs)
			if len(keys) > 0 {
				res.Files = append(res.Files, g.keysFile(m, majority, refs, keys))
			}
			if len(differing) > 0 {
				slices.Sort(differing)
				res.Diverging = append(res.Diverging, g.diverging(m, fmt.Sprintf("the keys %s differ from the siblings: they stay yours", strings.Join(differing, ", "))))
			}
		case whole != nil && m.exists:
			res.Diverging = append(res.Diverging, g.diverging(m, fmt.Sprintf("differs from the form that %d of %d sibling projects share%s: it stays yours", whole.n, len(voters), firstDifference(whole.rep, m))))
		}
	}
	return res
}

// keysFile is the keys-mode answer for a member. keys is nil for a file that is
// gone: then the keys come from the majority.
func (g *group) keysFile(m *member, majority map[string]string, refs References, keys map[string]any) protocol.DerivedFile {
	if keys == nil {
		keys, _ = renderKeys(majority, m, refs)
	}
	return protocol.DerivedFile{Project: m.project.Name, Path: m.rel, Mode: protocol.DeriveKeys, Format: "json", Renderer: Renderer, Version: Version, Keys: keys}
}

func (g *group) diverging(m *member, detail string) protocol.DivergentFile {
	return protocol.DivergentFile{Project: m.project.Name, Path: m.rel, Renderer: Renderer, Detail: "diverges from siblings: " + detail}
}

// firstDifference says where a text file first differs from the shape, as " (line
// 7: ...)"; it is empty for files that are not text.
func firstDifference(rep, m *member) string {
	if rep.isJSON || m.isJSON {
		return ""
	}
	want, have := strings.Split(rep.text, "\n"), strings.Split(m.text, "\n")
	for i := 0; i < len(want) || i < len(have); i++ {
		var w, h string
		if i < len(want) {
			w = want[i]
		}
		if i < len(have) {
			h = have[i]
		}
		if w != h {
			return fmt.Sprintf(" (line %d: %s)", i+1, strings.TrimSpace(tokensOf(m.project).concrete(h)))
		}
	}
	return ""
}
