package clicontract

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func tree() *cobra.Command {
	root := &cobra.Command{Use: "smith"}
	root.AddGroup(&cobra.Group{ID: "start", Title: "Create:"}, &cobra.Group{ID: "work", Title: "Work:"}, &cobra.Group{ID: "inspect", Title: "Inspect:"})
	add := &cobra.Command{Use: "add <kind> <name>", Short: "Add a project", GroupID: "start", Run: func(*cobra.Command, []string) {}}
	add.Flags().StringP("lang", "l", "", "language plugin")
	add.Flags().StringSlice("tag", nil, "tags, repeatable")
	add.Flags().Bool("dry-run", false, "change nothing")
	add.Flags().Int("parallel", 4, "tasks at once")
	add.Flags().Bool("no-install", false, "skip installing")
	build := &cobra.Command{Use: "build [project...]", Short: "Build projects", GroupID: "work", Run: func(*cobra.Command, []string) {}}
	doctor := &cobra.Command{Use: "doctor", Short: "Check the workspace", GroupID: "inspect", Run: func(*cobra.Command, []string) {}}
	hidden := &cobra.Command{Use: "serve", Hidden: true, Run: func(*cobra.Command, []string) {}}
	root.AddCommand(add, build, doctor, hidden)
	return root
}

func TestDescribeCommands(t *testing.T) {
	d := DescribeCommands(tree())
	names := []string{}
	for _, c := range d {
		names = append(names, c.Name)
	}
	if !reflect.DeepEqual(names, []string{"add", "build", "doctor"}) {
		t.Fatalf("commands %v", names)
	}
	add, build, doctor := d[0], d[1], d[2]
	if add.Group != "projects" || build.Group != "projects" || doctor.Group != "workspace" {
		t.Fatalf("groups %s %s %s", add.Group, build.Group, doctor.Group)
	}
	if !reflect.DeepEqual(add.Arguments, []ArgumentDescription{{Name: "kind", Description: "kind", Required: true}, {Name: "name", Description: "name", Required: true}}) {
		t.Fatalf("add arguments %+v", add.Arguments)
	}
	if !reflect.DeepEqual(build.Arguments, []ArgumentDescription{{Name: "project", Description: "project", Variadic: true}}) {
		t.Fatalf("build arguments %+v", build.Arguments)
	}
	opts := map[string]OptionDescription{}
	for _, o := range add.Options {
		opts[o.Name] = o
	}
	if o := opts["lang"]; o.Flags != "-l, --lang <lang>" || o.Short != "-l" || o.TakesValue != "required" {
		t.Errorf("lang %+v", o)
	}
	if o := opts["tag"]; !o.Repeatable {
		t.Errorf("tag %+v", o)
	}
	if o := opts["dry-run"]; o.Flags != "--dry-run" || o.TakesValue != "none" || o.DefaultValue != nil {
		t.Errorf("dry-run %+v", o)
	}
	if o := opts["parallel"]; o.DefaultValue != 4 {
		t.Errorf("parallel %+v", o)
	}
	if o := opts["no-install"]; !o.Negated {
		t.Errorf("no-install %+v", o)
	}
}

// The JSON keys are the ones the VS Code extension's CommandDescription,
// ArgumentDescription and OptionDescription interfaces read.
func TestCommandJSONKeepsTheExtensionsFieldNames(t *testing.T) {
	raw, _ := json.Marshal(DescribeCommands(tree()))
	var decoded []map[string]any
	_ = json.Unmarshal(raw, &decoded)
	keys := func(m map[string]any) []string {
		var out []string
		for k := range m {
			out = append(out, k)
		}
		sort.Strings(out)
		return out
	}
	if got := keys(decoded[0]); !reflect.DeepEqual(got, []string{"aliases", "arguments", "description", "group", "name", "options"}) {
		t.Fatalf("command keys %v", got)
	}
	arg := decoded[0]["arguments"].([]any)[0].(map[string]any)
	if got := keys(arg); !reflect.DeepEqual(got, []string{"description", "name", "required", "variadic"}) {
		t.Fatalf("argument keys %v", got)
	}
	for _, o := range decoded[0]["options"].([]any) {
		m := o.(map[string]any)
		for _, k := range []string{"name", "flags", "description", "takesValue", "variadic", "repeatable", "negated"} {
			if _, ok := m[k]; !ok {
				t.Fatalf("option %v lacks %s", m, k)
			}
		}
	}
}

func TestDescribeCommandsDescribesSubcommandsInTheirParentsGroup(t *testing.T) {
	root := &cobra.Command{Use: "smith"}
	root.AddGroup(&cobra.Group{ID: "work", Title: "Work:"})
	deps := &cobra.Command{Use: "deps", Short: "Dependencies", GroupID: "work"}
	up := &cobra.Command{Use: "up [packages...]", Short: "Update dependencies", Run: func(*cobra.Command, []string) {}}
	up.Flags().Bool("align", false, "one version")
	check := &cobra.Command{Use: "check", Short: "Check", Run: func(*cobra.Command, []string) {}}
	secret := &cobra.Command{Use: "secret", Hidden: true, Run: func(*cobra.Command, []string) {}}
	deps.AddCommand(up, check, secret)
	root.AddCommand(deps)

	d := DescribeCommands(root)
	if len(d) != 1 || len(d[0].Subcommands) != 2 {
		t.Fatalf("%+v", d)
	}
	if got := d[0].Subcommands[0]; got.Name != "check" || got.Group != "projects" {
		t.Errorf("sorted by name, in the parent's group: %+v", got)
	}
	got := d[0].Subcommands[1]
	if got.Name != "up" || len(got.Arguments) != 1 || got.Arguments[0].Name != "packages" || !got.Arguments[0].Variadic || len(got.Options) != 1 || got.Options[0].Name != "align" {
		t.Errorf("up: %+v", got)
	}
	raw, _ := json.Marshal(DescribeCommands(tree()))
	if strings.Contains(string(raw), "subcommands") {
		t.Error("a command without subcommands must not carry the key")
	}
}

func TestArgumentsOfTheUsageLinesThatWereReadWrong(t *testing.T) {
	for use, want := range map[string][]ArgumentDescription{
		"why-miss <project>:<task>":               {{Name: "project:task", Description: "project:task", Required: true}},
		"explain <project>[:<task>]":              {{Name: "project[:task]", Description: "project[:task]", Required: true}},
		"build [project]...":                      {{Name: "project", Description: "project", Variadic: true}},
		"set <name>...":                           {{Name: "name", Description: "name", Required: true, Variadic: true}},
		"check <project directory>...":            {{Name: "project directory", Description: "project directory", Required: true, Variadic: true}},
		"set <setting> <value...>":                {{Name: "setting", Description: "setting", Required: true}, {Name: "value", Description: "value", Required: true, Variadic: true}},
		"check <name | path to plugin> [args...]": {{Name: "name | path to plugin", Description: "name | path to plugin", Required: true}},
	} {
		if got := arguments("cmd " + use); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %+v, want %+v", use, got, want)
		}
	}
}
