package clicontract

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// groups maps the help's groups onto the editor's: creating and checking a
// workspace, working on projects, shipping, and inspecting.
var groups = map[string]string{"start": "workspace", "work": "projects", "ship": "pipeline", "inspect": "inspect"}

// regrouped are commands whose editor group differs from their help group,
// as the TypeScript smith grouped them.
var regrouped = map[string]string{"doctor": "workspace", "add": "projects", "remove": "projects", "install": "dependencies", "uninstall": "dependencies", "sync": "dependencies", "up": "dependencies"}

var argPattern = regexp.MustCompile(`<([^>]+)>|\[([^\]]+)\]`)

// DescribeCommands describes root's visible commands, sorted by name, each with
// its own visible subcommands.
func DescribeCommands(root *cobra.Command) []CommandDescription {
	var out []CommandDescription
	for _, c := range root.Commands() {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		group := groups[c.GroupID]
		if group == "" {
			group = "inspect"
		}
		if g, ok := regrouped[c.Name()]; ok {
			group = g
		}
		out = append(out, describe(c, group))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// describe describes one command and, under it, its visible subcommands, which
// are in the group of their parent.
func describe(c *cobra.Command, group string) CommandDescription {
	d := CommandDescription{
		Name: c.Name(), Aliases: append([]string{}, c.Aliases...), Group: group,
		Description: c.Short, Arguments: arguments(c.Use), Options: []OptionDescription{},
	}
	c.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden || f.Name == "help" {
			return
		}
		d.Options = append(d.Options, option(f))
	})
	for _, sub := range c.Commands() {
		if sub.Hidden || sub.Name() == "help" || sub.Name() == "completion" {
			continue
		}
		d.Subcommands = append(d.Subcommands, describe(sub, group))
	}
	sort.Slice(d.Subcommands, func(i, j int) bool { return d.Subcommands[i].Name < d.Subcommands[j].Name })
	return d
}

func arguments(use string) []ArgumentDescription {
	_, rest, _ := strings.Cut(use, " ")
	out := []ArgumentDescription{}
	for _, m := range argPattern.FindAllStringSubmatch(rest, -1) {
		name, required := m[1], true
		if name == "" {
			name, required = m[2], false
		}
		variadic := strings.HasSuffix(name, "...")
		name = strings.TrimSuffix(name, "...")
		if strings.HasPrefix(name, ":") || name == "args" || name == "flags" {
			continue
		}
		out = append(out, ArgumentDescription{Name: name, Description: name, Required: required, Variadic: variadic})
	}
	return out
}

func option(f *pflag.Flag) OptionDescription {
	o := OptionDescription{Name: f.Name, Description: f.Usage, TakesValue: "required", Negated: strings.HasPrefix(f.Name, "no-")}
	switch f.Value.Type() {
	case "bool":
		o.TakesValue = "none"
	case "stringArray", "stringSlice":
		o.Repeatable = true
	}
	o.Flags = "--" + f.Name
	if f.Shorthand != "" {
		o.Short = "-" + f.Shorthand
		o.Flags = o.Short + ", " + o.Flags
	}
	if o.TakesValue != "none" {
		o.Flags += " <" + f.Name + ">"
	}
	o.DefaultValue = defaultValue(f)
	return o
}

func defaultValue(f *pflag.Flag) any {
	switch v := f.DefValue; {
	case v == "" || v == "[]" || v == "false" || (v == "0" && f.Value.Type() == "int"):
		return nil
	case f.Value.Type() == "int":
		n, _ := strconv.Atoi(v)
		return n
	case f.Value.Type() == "bool":
		return v == "true"
	default:
		return v
	}
}
