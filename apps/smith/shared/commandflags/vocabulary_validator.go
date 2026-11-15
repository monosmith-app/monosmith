package commandflags

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Violations walks the whole command tree under root and returns one sentence
// for each flag that breaks the flag vocabulary (ADR 0084), sorted. It reads
// the tree in memory: it opens no session and starts no plugin.
//
// The rules: (1) no command defines, in plain view, a near-synonym or a synonym
// of a shared concept under another name; (2) a shared flag carries the
// vocabulary's shorthand and every synonym; (3) a concept that lists its commands
// is taken by exactly those commands; (4) one shorthand never stands for two
// flag names. A flag in flagvocabulary.Exceptions is excused from rules 1 and 4.
//
// Parameters: root is the command tree, as the program builds it.
func Violations(root *cobra.Command) []string {
	excused := map[string]bool{}
	for _, e := range flagvocabulary.Exceptions() {
		excused[e.Command+"|"+e.Flag] = true
	}
	var out []string
	shorthands := map[string]map[string]bool{}
	has := map[string][]string{}
	walkCommands(root, func(path string, cmd *cobra.Command) {
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Name == "help" {
				return
			}
			skip := excused[path+"|"+f.Name]
			if !f.Hidden && f.Shorthand != "" && !skip {
				if shorthands[f.Shorthand] == nil {
					shorthands[f.Shorthand] = map[string]bool{}
				}
				shorthands[f.Shorthand][f.Name] = true
			}
			for _, c := range flagvocabulary.All() {
				switch {
				case f.Name == c.Flag:
					has[c.ID] = append(has[c.ID], path)
					out = append(out, checkShared(path, cmd, f, c)...)
				case !f.Hidden && !skip && isNear(f.Name, c):
					out = append(out, fmt.Sprintf("smith %s: --%s is another name for %s: use --%s (add it through package commandflags)", path, f.Name, c.Meaning, c.Flag))
				}
			}
		})
	})
	for _, c := range flagvocabulary.All() {
		if len(c.Commands) == 0 {
			continue
		}
		got := has[c.ID]
		sort.Strings(got)
		for _, p := range got {
			if !slices.Contains(c.Commands, p) {
				out = append(out, fmt.Sprintf("smith %s takes --%s but flag_vocabulary.json does not list it under %q", p, c.Flag, c.ID))
			}
		}
		for _, p := range c.Commands {
			if !slices.Contains(got, p) {
				out = append(out, fmt.Sprintf("flag_vocabulary.json lists smith %s under %q but the command has no --%s", p, c.ID, c.Flag))
			}
		}
	}
	for s, names := range shorthands {
		if len(names) > 1 {
			var list []string
			for n := range names {
				list = append(list, "--"+n)
			}
			sort.Strings(list)
			out = append(out, fmt.Sprintf("-%s stands for more than one flag (%s): one shorthand, one meaning", s, strings.Join(list, ", ")))
		}
	}
	sort.Strings(out)
	return out
}

// checkShared checks one use of a shared flag: its shorthand and its synonyms.
func checkShared(path string, cmd *cobra.Command, f *pflag.Flag, c flagvocabulary.Concept) []string {
	var out []string
	if f.Shorthand != c.Short {
		out = append(out, fmt.Sprintf("smith %s: --%s has shorthand %q, the vocabulary says %q", path, f.Name, f.Shorthand, c.Short))
	}
	for _, s := range c.Synonyms {
		syn := cmd.Flags().Lookup(s.Name)
		if syn == nil || syn.Shorthand != s.Short {
			out = append(out, fmt.Sprintf("smith %s: --%s is missing its synonym --%s: register it through package commandflags", path, f.Name, s.Name))
		}
	}
	return out
}

// isNear reports whether name is a near-synonym of the concept, or one of its
// synonyms in plain view (a synonym must be hidden: the help names it once).
func isNear(name string, c flagvocabulary.Concept) bool {
	if slices.Contains(c.Near, name) {
		return true
	}
	for _, s := range c.Synonyms {
		if s.Name == name {
			return true
		}
	}
	return false
}

// walkCommands calls visit for root's descendants (not root itself), with the
// command path under smith ("release assets").
func walkCommands(root *cobra.Command, visit func(path string, cmd *cobra.Command)) {
	var walk func(c *cobra.Command, prefix string)
	walk = func(c *cobra.Command, prefix string) {
		for _, sub := range c.Commands() {
			if sub.Name() == "help" || sub.Name() == "completion" {
				continue
			}
			path := strings.TrimSpace(prefix + " " + sub.Name())
			visit(path, sub)
			walk(sub, path)
		}
	}
	walk(root, "")
}
