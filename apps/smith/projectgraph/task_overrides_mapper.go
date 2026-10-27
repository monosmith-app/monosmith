package projectgraph

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/splitrepo"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TaskFields are the fields of a task whose source is tracked.
var TaskFields = []string{"command", "cwd", "inputs", "outputs", "dependsOn", "env", "cache"}

// ApplyOverrides layers smith.toml on top of plugin-inferred tasks: plugin
// defaults, then [tasks.<task>], then [projects.<name>.tasks.<task>]. A
// project override may also add a task the plugin does not know about, if it
// gives a command. Every field's source is recorded in Sources.
func ApplyOverrides(p *Project, cfg workspace.Config) error {
	p.Sources = map[string]string{}
	if p.Tasks == nil {
		p.Tasks = map[string]protocol.Task{}
	}
	for name := range p.Tasks {
		for _, f := range TaskFields {
			p.Sources[name+"."+f] = "plugin " + p.Plugin
		}
	}
	for field, source := range p.contributed {
		p.Sources[field] = source
	}
	// An app's package task runs in smith ci verify, so a packaging that breaks
	// shows on the pull request and not at the release; verify = false turns
	// it off.
	if t, ok := p.Tasks["package"]; ok && p.Kind == protocol.ProjectApp && len(t.Command) > 0 {
		t.Verify = true
		p.Tasks["package"] = t
	}
	if g := cfg.Projects[p.Name].Generate; g != nil {
		if err := p.addGenerate(*g); err != nil {
			return fmt.Errorf("smith.toml [projects.%s.generate]: %w", p.Name, err)
		}
	}
	names := make([]string, 0, len(p.Tasks))
	for name := range p.Tasks {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if o, ok := cfg.Tasks[name]; ok {
			p.override(name, o, fmt.Sprintf("smith.toml [tasks.%s]", name))
		}
	}
	p.Registries = ProjectRegistries(cfg, p.Name)
	if po, ok := cfg.Projects[p.Name]; ok {
		if po.TagPrefix != "" {
			p.TagPrefix = po.TagPrefix
		}
		p.TagPrefixSet = po.TagPrefix != ""
		for _, tag := range po.Tags {
			if !slices.Contains(p.Tags, tag) {
				p.Tags = append(p.Tags, tag)
			}
		}
		sort.Strings(p.Tags)
		p.BareTags = po.BareTags
		p.Release = po.Release
		p.ReleaseAssets =p.AssetSpec()
		if po.SplitRepo != "" {
			if err := splitrepo.Valid(po.SplitRepo); err != nil {
				return fmt.Errorf("smith.toml [projects.%s]: %w", p.Name, err)
			}
			p.SplitRepo = po.SplitRepo
		}
		taskNames := make([]string, 0, len(po.Tasks))
		for name := range po.Tasks {
			taskNames = append(taskNames, name)
		}
		sort.Strings(taskNames)
		for _, name := range taskNames {
			o := po.Tasks[name]
			t, exists := p.Tasks[name]
			if !exists && len(o.Command) == 0 {
				return fmt.Errorf("smith.toml [projects.%s.tasks.%s]: %s has no %q task, so the override must give a command", p.Name, name, p.Name, name)
			}
			p.Tasks[name] = t
			p.override(name, o, fmt.Sprintf("smith.toml [projects.%s.tasks.%s]", p.Name, name))
		}
	}
	return nil
}

// override applies one smith.toml override to a task. A depends_on override
// replaces the list, but a dependency on a task the same plugin contributed
// to the project stays: such a task (e2e-browsers before e2e) exists only for
// the one that needs it, so dropping it breaks the task rather than changing
// it. Any other dependency a plugin added and the override leaves out is
// recorded in Dropped, for doctor to name.
func (p *Project) override(name string, o workspace.TaskOverride, source string) {
	before := p.Tasks[name]
	from := p.Sources[name+".dependsOn"]
	t := apply(before, o, p.Sources, name, source)
	// A plugin added the list: a language plugin, a stack's toolchain ("plugin
	// x (stack y)") or a convention ("convention x"), not an earlier override.
	if o.DependsOn != nil && from != "" && !strings.HasPrefix(from, "smith.toml") {
		for _, d := range before.DependsOn {
			if slices.Contains(t.DependsOn, d) {
				continue
			}
			if p.contributed[name+".dependsOn"] == from && p.contributed[d+".command"] == from {
				t.DependsOn = append(slices.Clone(t.DependsOn), d)
				p.Sources[name+".dependsOn"] = source + ", with " + d + " from " + from
				continue
			}
			if p.Dropped == nil {
				p.Dropped = map[string][]string{}
			}
			p.Dropped[name] = append(p.Dropped[name], d+" (from "+from+")")
		}
	}
	p.Tasks[name] = t
}

func apply(t protocol.Task, o workspace.TaskOverride, sources map[string]string, name, source string) protocol.Task {
	set := func(field string) { sources[name+"."+field] = source }
	if o.Command != nil {
		t.Command = o.Command
		set("command")
	}
	if o.Cwd != "" {
		t.Cwd = o.Cwd
		set("cwd")
	}
	if o.Inputs != nil {
		t.Inputs = o.Inputs
		set("inputs")
	}
	if o.Outputs != nil {
		t.Outputs = o.Outputs
		set("outputs")
	}
	if o.DependsOn != nil {
		t.DependsOn = o.DependsOn
		set("dependsOn")
	}
	if o.Env != nil {
		t.Env = o.Env
		set("env")
	}
	if o.Cache != nil {
		t.Cache = *o.Cache
		set("cache")
	}
	if o.Verify != nil {
		t.Verify = *o.Verify
	}
	if o.Serve != "" {
		t.Serve = o.Serve
		set("serve")
	}
	if o.ServeURL != "" {
		t.ServeURL = o.ServeURL
		set("serve_url")
	}
	if o.ServeTimeout != "" {
		t.ServeTimeout = o.ServeTimeout
		set("serve_timeout")
	}
	return t
}
