package choicedata

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Names of the files in a choices directory that are not choices.
const (
	catalogName = "catalog.toml"
	baseName    = "base.toml"
	stacksName  = "stacks.toml"
)

// Set is the choices a plugin ships, checked: its choices, the base they
// build on, the stacks it publishes and the version catalog.
type Set struct {
	choices  []choice
	base     *choice
	stacks   []protocol.Stack
	versions map[string]string
}

// choice is a File with its roles converted and its name settled.
type choice struct {
	File
	roles []protocol.Role
}

// Load reads and checks the choices in dir of fsys (an embedded folder):
// catalog.toml (versions), base.toml, stacks.toml, and every other .toml file
// as a choice. Files of other kinds are ignored. An unknown member, a
// package with no version in the catalog, a name used twice, a file path that
// leaves the project, or a choice that protocol.ValidateChoice refuses is an
// error that names the file.
//
// An empty directory is a valid, empty Set.
func Load(fsys fs.FS, dir string) (*Set, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	s := &Set{versions: map[string]string{}}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".toml") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	for _, name := range files {
		raw, err := fs.ReadFile(fsys, path.Join(dir, name))
		if err != nil {
			return nil, err
		}
		if err := s.add(name, raw); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}
	return s, s.check()
}

// add reads one file into the set.
func (s *Set) add(name string, raw []byte) error {
	switch name {
	case catalogName:
		var f CatalogFile
		if err := decode(raw, &f); err != nil {
			return err
		}
		s.versions = f.Versions
	case stacksName:
		var f StacksFile
		if err := decode(raw, &f); err != nil {
			return err
		}
		for _, e := range f.Stack {
			s.stacks = append(s.stacks, protocol.Stack{Name: e.Name, Description: e.Description, Choices: e.Choices, Toolchain: e.Toolchain})
		}
	default:
		var f File
		if err := decode(raw, &f); err != nil {
			return err
		}
		c := choice{File: f}
		for _, r := range f.Roles {
			c.roles = append(c.roles, protocol.Role(r))
		}
		if name == baseName {
			s.base = &c
			return nil
		}
		if want := strings.TrimSuffix(name, ".toml"); f.Name != want {
			return fmt.Errorf("the choice is named %q, but the file must be named %s.toml", f.Name, f.Name)
		}
		s.choices = append(s.choices, c)
	}
	return nil
}

// decode reads TOML and refuses a member the format does not have.
func decode(raw []byte, into any) error {
	d := toml.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(into); err != nil {
		var strict *toml.StrictMissingError
		if errors.As(err, &strict) {
			return fmt.Errorf("unknown member: %s", strings.TrimSpace(strict.String()))
		}
		return err
	}
	return nil
}

// check applies the rules that need the whole set.
func (s *Set) check() error {
	seen := map[string]bool{}
	for _, c := range s.choices {
		if err := protocol.ValidateChoice(c.protocol()); err != nil {
			return err
		}
		if seen[c.Name] {
			return fmt.Errorf("the choice %s is defined twice", c.Name)
		}
		seen[c.Name] = true
	}
	all := s.choices
	if s.base != nil {
		all = append(append([]choice(nil), s.choices...), *s.base)
	}
	for _, c := range all {
		who := c.Name
		if c.Name == "" {
			who = "base"
		}
		for _, list := range [][]string{c.Dependencies, c.DevDependencies, c.Overrides} {
			for _, pkg := range list {
				if s.versions[pkg] == "" {
					return fmt.Errorf("%s names the package %s, which has no version in %s", who, pkg, catalogName)
				}
			}
		}
		for file := range c.Files {
			if clean := path.Clean(file); clean != file || path.IsAbs(file) || strings.HasPrefix(file, "../") || file == "." {
				return fmt.Errorf("%s: the file %q must be a path inside the project, in its clean form", who, file)
			}
		}
		for slot := range c.Slots {
			if slot == "" {
				return fmt.Errorf("%s: a slot has no name", who)
			}
		}
	}
	for _, st := range s.stacks {
		if st.Name == "" || len(st.Choices) == 0 {
			return fmt.Errorf("%s: a stack needs a name and choices", stacksName)
		}
	}
	return nil
}

// protocol is the choice as the plugin publishes it.
func (c choice) protocol() protocol.Choice {
	return protocol.Choice{
		Axis: c.Axis, Name: c.Name, Description: c.Description, Language: c.Language, Roles: c.roles,
		Needs: c.Needs, Default: c.Default, Many: c.Many, Requires: c.Requires, Conflicts: c.Conflicts,
	}
}

// Choices are the choices the set publishes (InitializeResult.Choices), sorted
// by axis and name.
func (s *Set) Choices() []protocol.Choice {
	out := make([]protocol.Choice, len(s.choices))
	for i, c := range s.choices {
		out[i] = c.protocol()
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Axis != out[j].Axis {
			return out[i].Axis < out[j].Axis
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// Stacks are the stacks the set publishes (InitializeResult.Stacks), each a
// saved selection of choices.
func (s *Set) Stacks() []protocol.Stack { return append([]protocol.Stack(nil), s.stacks...) }
