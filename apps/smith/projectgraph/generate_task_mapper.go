package projectgraph

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// GenerateTask is the name of the task that turns a shared schema into code.
const GenerateTask = "generate"

// generatedBefore are the tasks that read the generated code, so they wait
// for the generate task when the project has them.
var generatedBefore = []string{"lint", "typecheck", "test", "build"}

// addGenerate adds the generate task for [projects.<name>.generate] and
// orders it before the project's build, lint, typecheck and test. The task
// reads the schema (a workspace file, so cached on its content and the pinned
// generator version) and writes the code to its output directory.
func (p *Project) addGenerate(g workspace.GenerateConfig) error {
	task, err := GenerateTaskFor(p.Root, g)
	if err != nil {
		return err
	}
	if p.Tasks == nil {
		p.Tasks = map[string]protocol.Task{}
	}
	p.Tasks[GenerateTask] = task
	for _, f := range TaskFields {
		p.Sources[GenerateTask+"."+f] = "smith.toml [projects." + p.Name + ".generate]"
	}
	for _, name := range generatedBefore {
		t, ok := p.Tasks[name]
		if !ok {
			continue
		}
		already := false
		for _, d := range t.DependsOn {
			already = already || d == GenerateTask
		}
		if !already {
			t.DependsOn = append([]string{GenerateTask}, t.DependsOn...)
			p.Tasks[name] = t
			p.Sources[name+".dependsOn"] = "smith.toml [projects." + p.Name + ".generate]"
		}
	}
	return nil
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
		if g.Generator == "" {
			return protocol.Task{}, fmt.Errorf("generator is required unless command is given")
		}
		switch kind {
		case "openapi":
			command = []string{"npx", "--yes", "@openapitools/openapi-generator-cli@" + g.Version, "generate", "-i", "{schema}", "-g", "{generator}", "-o", "{out}"}
		case "protobuf":
			command = []string{"protoc", "-I", "{schema_dir}", "--{generator}_out={out}", "{schema}"}
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
