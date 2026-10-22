package projectgraph

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// GenerateTask is the name of the task that turns a shared schema into code.
const GenerateTask = "generate"

// generatedBefore are the tasks that read the generated code, so they wait
// for the generate tasks when the project has them.
var generatedBefore = []string{"lint", "typecheck", "test", "build"}

// IsGenerateTask says whether a task name is a generate task: "generate", or
// "generate-<name>" (one of several, or one a plugin contributes). Such a task
// runs before the tasks that read the code it writes.
func IsGenerateTask(name string) bool {
	return name == GenerateTask || strings.HasPrefix(name, GenerateTask+"-")
}

// generateEntry is one generate task a project's smith.toml asks for.
type generateEntry struct {
	task   string
	config workspace.GenerateConfig
}

var nonNameChars = regexp.MustCompile(`[^a-z0-9]+`)

// generateEntries lists the tasks [projects.<name>.generate] and
// [[projects.<name>.generates]] stand for: "generate" for the table, and
// "generate-<name>" for each array entry. Output directories must not overlap,
// since a task clears its own before it runs.
func generateEntries(po workspace.ProjectOverride) ([]generateEntry, error) {
	var entries []generateEntry
	if po.Generate != nil {
		entries = append(entries, generateEntry{GenerateTask, *po.Generate})
	}
	seen := map[string]bool{}
	for i, g := range po.Generates {
		name := g.Name
		if name == "" {
			name = strings.TrimSuffix(path.Base(g.Schema), path.Ext(g.Schema))
		}
		name = strings.Trim(nonNameChars.ReplaceAllString(strings.ToLower(name), "-"), "-")
		if name == "" {
			return nil, fmt.Errorf("generates[%d]: give a name or a schema", i)
		}
		if seen[name] {
			return nil, fmt.Errorf("generates[%d]: two entries are named %q: give each a name", i, name)
		}
		seen[name] = true
		if g.Out == "" {
			g.Out = "generated-" + name
		}
		entries = append(entries, generateEntry{GenerateTask + "-" + name, g})
	}
	outs := make([]string, len(entries))
	for i, e := range entries {
		out := e.config.Out
		if out == "" {
			out = "generated"
		}
		outs[i] = path.Clean(strings.ReplaceAll(out, `\`, "/"))
		for j := 0; j < i; j++ {
			if outs[i] == outs[j] || strings.HasPrefix(outs[i], outs[j]+"/") || strings.HasPrefix(outs[j], outs[i]+"/") {
				return nil, fmt.Errorf("%s and %s write to overlapping directories (%s, %s): give each its own out", entries[j].task, e.task, outs[j], outs[i])
			}
		}
	}
	return entries, nil
}

// addGenerate adds the generate tasks for [projects.<name>.generate] and
// [[projects.<name>.generates]]. Each reads its schema (a workspace file, so
// cached on its content and the pinned generator version) and writes the code
// to its output directory.
func (p *Project) addGenerate(po workspace.ProjectOverride) error {
	entries, err := generateEntries(po)
	if err != nil {
		return err
	}
	if p.Tasks == nil {
		p.Tasks = map[string]protocol.Task{}
	}
	for _, e := range entries {
		task, err := GenerateTaskFor(p.Root, e.config)
		if err != nil {
			return fmt.Errorf("%s: %w", e.task, err)
		}
		p.Tasks[e.task] = task
		for _, f := range TaskFields {
			p.Sources[e.task+"."+f] = "smith.toml [projects." + p.Name + ".generate]"
		}
	}
	p.orderGenerateTasks()
	return nil
}

// orderGenerateTasks puts every generate task of the project, whoever added
// it (smith.toml, or a plugin's contribution), before the project's lint,
// typecheck, test and build.
func (p *Project) orderGenerateTasks() {
	var generates []string
	for name := range p.Tasks {
		if IsGenerateTask(name) {
			generates = append(generates, name)
		}
	}
	sort.Strings(generates)
	if len(generates) == 0 {
		return
	}
	if p.Sources == nil {
		p.Sources = map[string]string{}
	}
	for _, name := range generatedBefore {
		t, ok := p.Tasks[name]
		if !ok {
			continue
		}
		var missing []string
		for _, g := range generates {
			if !slices.Contains(t.DependsOn, g) {
				missing = append(missing, g)
			}
		}
		if len(missing) > 0 {
			t.DependsOn = append(missing, t.DependsOn...)
			p.Tasks[name] = t
			p.Sources[name+".dependsOn"] = "generate tasks of " + p.Name
		}
	}
}

// GenerateTaskFor builds the generate task of a project rooted at root.
func GenerateTaskFor(root string, g workspace.GenerateConfig) (protocol.Task, error) {
	schema := path.Clean(strings.ReplaceAll(g.Schema, `\`, "/"))
	kind := SchemaKind(schema)
	if g.Schema == "" {
		return protocol.Task{}, fmt.Errorf("schema is required")
	}
	if g.Version == "" {
		return protocol.Task{}, fmt.Errorf("version is required: pin the generator so a new version regenerates, not the machine's")
	}
	out := g.Out
	if out == "" {
		out = "generated"
	}
	out = path.Clean(strings.ReplaceAll(out, `\`, "/"))
	if out == "." || out == ".." || strings.HasPrefix(out, "../") || path.IsAbs(out) {
		return protocol.Task{}, fmt.Errorf("out %q must be a directory inside the project", g.Out)
	}
	rel, err := filepath.Rel(filepath.FromSlash(root), filepath.FromSlash(schema))
	if err != nil {
		return protocol.Task{}, err
	}
	rel = filepath.ToSlash(rel)
	command := g.Command
	if len(command) == 0 {
		switch {
		case kind == "graphql" && g.Generator != "" && g.Generator != "typescript":
			return protocol.Task{}, fmt.Errorf("the default GraphQL generator writes a TypeScript client (generator \"typescript\", or none), not %q: give command", g.Generator)
		case kind != "graphql" && g.Generator == "":
			return protocol.Task{}, fmt.Errorf("generator is required unless command is given")
		}
		switch kind {
		case "openapi":
			command = []string{"npx", "--yes", "@openapitools/openapi-generator-cli@" + g.Version, "generate", "-i", "{schema}", "-g", "{generator}", "-o", "{out}"}
		case "protobuf":
			command = []string{"protoc", "-I", "{schema_dir}", "--{generator}_out={out}", "{schema}"}
		case "asyncapi":
			command = []string{"npx", "--yes", "@asyncapi/cli@" + g.Version, "generate", "fromTemplate", "{schema}", "{generator}", "-o", "{out}", "--force-write"}
		case "graphql":
			command = []string{"npx", "--yes", "@genql/cli@" + g.Version, "--schema", "{schema}", "--output", "{out}"}
		default:
			return protocol.Task{}, fmt.Errorf("no default generator for %q: give command", g.Schema)
		}
	}
	replacer := strings.NewReplacer("{schema_dir}", path.Dir(rel), "{schema}", rel, "{out}", out, "{generator}", g.Generator)
	argv := make([]string, len(command))
	for i, a := range command {
		argv[i] = replacer.Replace(a)
	}
	return protocol.Task{
		Command:     argv,
		Env:         map[string]string{"SMITH_GENERATOR_VERSION": g.Version},
		Inputs:      []string{"{workspaceRoot}/" + schema},
		Outputs:     []string{out},
		Cache:       true,
		Description: "Generate code from " + schema,
	}, nil
}
