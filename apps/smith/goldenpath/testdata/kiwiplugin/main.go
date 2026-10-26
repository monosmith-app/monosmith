// Command kiwiplugin is an external smith plugin for "kiwi", a language smith
// knows nothing about, used by the end-to-end tests to prove a language can
// be added by a plugin alone. A kiwi project is a directory with a kiwi.toml:
//
//	name = "fruit"
//	kind = "package"
//	version = "0.0.0"
//	imports = "kiwi:other, kiwi:more"
//
// The same binary is kiwi's build tool: `kiwiplugin build|test|fmt` runs in a
// project directory. It plays two roles: language, and the toolchain of the
// kiwi-strict stack it publishes, which adds fmt and makes test strict.
package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

func main() {
	if len(os.Args) > 1 {
		os.Exit(tool(os.Args[1:]))
	}
	if err := serve.Serve(&kiwi{}, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// tool is kiwi's build tool.
func tool(args []string) int {
	switch args[0] {
	case "build":
		// Into the directory smith gives it: the workspace's dist/<project dir>.
		out := args[1]
		if err := os.MkdirAll(out, 0o755); err != nil {
			return 1
		}
		if err := os.WriteFile(filepath.Join(out, "built.txt"), []byte("kiwi\n"), 0o644); err != nil {
			return 1
		}
		fmt.Println("kiwi built")
	case "test":
		fmt.Println("kiwi tests passed", strings.Join(args[1:], " "))
	case "fmt":
		fmt.Println("kiwi formatted")
	case "greet":
		fmt.Println("kiwi says", strings.Join(args[1:], " "))
	default:
		return 2
	}
	return 0
}

type kiwi struct{ root string }

func self() string {
	exe, _ := os.Executable()
	return filepath.ToSlash(exe)
}

func (k *kiwi) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	k.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: "kiwi", Version: "1.0.0"},
		Kind:      protocol.KindLanguage,
		Languages: []string{"kiwi"},
		Manifests: []string{"kiwi.toml"},
		Roles:     []protocol.Role{protocol.RoleLanguage, protocol.RoleToolchain, protocol.RoleScaffold},
		Stacks: []protocol.Stack{{
			Name: "kiwi-strict", Description: "kiwi with formatting and strict tests",
			Language: "kiwi", Toolchain: []string{"kiwi"},
		}},
	}, nil
}

// manifest reads kiwi.toml's "key = value" lines.
func (k *kiwi) manifest(rel string) (map[string]string, error) {
	raw, err := os.ReadFile(filepath.Join(k.root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, err
	}
	m := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		if key, value, ok := strings.Cut(line, "="); ok {
			m[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"`)
		}
	}
	return m, nil
}

func (k *kiwi) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	var res protocol.DiscoverResult
	for _, rel := range params.Manifests {
		m, err := k.manifest(rel)
		if err != nil {
			return res, err
		}
		p := protocol.Project{
			Name: m["name"], Root: path.Dir(rel), Language: "kiwi", Kind: protocol.ProjectKind(m["kind"]),
			Version: m["version"], Stack: m["stack"], Exports: []string{"kiwi:" + m["name"]},
			Tasks: map[string]protocol.Task{
				"build": {Command: []string{self(), "build", "{workspaceRoot}/" + protocol.ProjectBuildDir(path.Dir(rel))}, Outputs: []string{protocol.BuildDirFromProject(path.Dir(rel))}, DependsOn: []string{"^build"}, Cache: true},
				"test":  {Command: []string{self(), "test"}, DependsOn: []string{"build"}, Cache: true},
			},
		}
		for _, imp := range strings.Split(m["imports"], ",") {
			if imp = strings.TrimSpace(imp); imp != "" {
				p.Imports = append(p.Imports, imp)
			}
		}
		res.Projects = append(res.Projects, p)
	}
	return res, nil
}

func (k *kiwi) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(k.root, filepath.FromSlash(params.Directory))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return protocol.GenerateResult{}, err
	}
	body := fmt.Sprintf("name = %q\nkind = %q\nversion = \"0.0.0\"\n", params.Name, params.Kind)
	if params.Stack != "" {
		body += fmt.Sprintf("stack = %q\n", params.Stack)
	}
	if err := os.WriteFile(filepath.Join(dir, "kiwi.toml"), []byte(body), 0o644); err != nil {
		return protocol.GenerateResult{}, err
	}
	files := []string{params.Directory + "/kiwi.toml"}
	// A recipe's settings for kiwi arrive in the generate call: kiwi writes
	// the greeting it was given where the end-to-end test can read it.
	if greeting, ok := params.Settings["greeting"].(string); ok {
		if err := os.WriteFile(filepath.Join(dir, "greeting.txt"), []byte(greeting+"\n"), 0o644); err != nil {
			return protocol.GenerateResult{}, err
		}
		files = append(files, params.Directory+"/greeting.txt")
		// kiwi also writes a file it owns, which doctor restores from the
		// setting: the one an upgrade rewrites. greeting.txt is the starter.
		if err := os.WriteFile(filepath.Join(dir, ownedFile), []byte(owned(greeting)), 0o644); err != nil {
			return protocol.GenerateResult{}, err
		}
		files = append(files, params.Directory+"/"+ownedFile)
	}
	return protocol.GenerateResult{Files: files}, nil
}

// ownedFile is the file kiwi owns in a project that has a greeting.
const ownedFile = "kiwi.owned.txt"

// owned is the content kiwi owns in ownedFile for a greeting.
func owned(greeting string) string {
	return "# Generated by kiwi from the greeting setting; kiwi rewrites it.\ngreeting = " + greeting + "\n"
}

// Contribute is the kiwi-strict stack's toolchain: fmt, and strict tests.
func (k *kiwi) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, p := range params.Projects {
		c := protocol.Contribution{
			Tasks:     map[string]protocol.Task{"fmt": {Command: []string{self(), "fmt"}, Cache: true}},
			Replace:   map[string]protocol.Task{"test": {Command: []string{self(), "test", "--strict"}, DependsOn: []string{"build"}, Cache: true}},
			DependsOn: map[string][]string{"test": {"fmt"}},
		}
		// The settings smith.toml holds for kiwi reach it with every project:
		// a greeting becomes a task that says it.
		if greeting, ok := p.Settings["kiwi"]["greeting"].(string); ok {
			c.Tasks["greet"] = protocol.Task{Command: []string{self(), "greet", greeting}}
		}
		// A schema setting becomes a generate task: kiwi names the schema it
		// reads, and smith orders it before the project's build and links the
		// project to the schema's owner.
		if schema, ok := p.Settings["kiwi"]["schema"].(string); ok {
			c.Tasks["generate-schema"] = protocol.Task{
				Command:     []string{"protoc", "-I", "{workspaceRoot}/" + path.Dir(schema), "--python_out=gen", "{workspaceRoot}/" + schema},
				Inputs:      []string{"{workspaceRoot}/" + schema},
				Outputs:     []string{"gen"},
				Cache:       true,
				Description: "Generate code from " + schema,
			}
		}
		res.Projects[p.Name] = c
	}
	return res, nil
}

func (k *kiwi) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	rel := params.Project.Root + "/kiwi.toml"
	if params.DryRun {
		return protocol.VersionResult{Files: []string{rel}}, nil
	}
	p := filepath.Join(k.root, filepath.FromSlash(rel))
	raw, err := os.ReadFile(p)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	var out []string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "version") {
			line = fmt.Sprintf("version = %q", params.Version)
		}
		out = append(out, line)
	}
	return protocol.VersionResult{Files: []string{rel}}, os.WriteFile(p, []byte(strings.Join(out, "\n")), 0o644)
}

// Publish appends to a registry file at the workspace root, skipping a
// version already there.
func (k *kiwi) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	entry := params.Project.Name + "@" + params.Version
	registry := filepath.Join(k.root, ".kiwi-registry")
	raw, _ := os.ReadFile(registry)
	if strings.Contains(string(raw), entry+"\n") {
		return protocol.PublishResult{Skipped: entry + " is already published"}, nil
	}
	if params.DryRun {
		return protocol.PublishResult{Published: true, Registry: "kiwi"}, nil
	}
	f, err := os.OpenFile(registry, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return protocol.PublishResult{}, err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(entry + "\n")
	return protocol.PublishResult{Published: true, Registry: "kiwi"}, err
}

func (k *kiwi) CI(params protocol.CIParams) (protocol.CIResult, error) {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}, nil
	}
	return protocol.CIResult{Toolchains: []protocol.Toolchain{{Name: "kiwi", Version: "1.0", Install: "curl -fsSL https://example.com/install-kiwi.sh | sh"}}}, nil
}

func (k *kiwi) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	var res protocol.DoctorResult
	for _, p := range params.Projects {
		res.Findings = append(res.Findings, protocol.Finding{Check: "kiwi project " + p.Name + " has a kiwi.toml", OK: true})
		greeting, ok := p.Settings["kiwi"]["greeting"].(string)
		if !ok {
			continue
		}
		// The owned file follows the greeting setting; a safe fix restores it.
		rel := p.Root + "/" + ownedFile
		want := owned(greeting)
		got, _ := os.ReadFile(filepath.Join(k.root, filepath.FromSlash(rel)))
		finding := protocol.Finding{
			Check: "kiwi project " + p.Name + " owned file matches its greeting", OK: string(got) == want,
			Detail: rel + " is not what the greeting setting says", Remedy: "rewrite it from the setting",
		}
		if !finding.OK {
			finding.Fix = &protocol.Fix{Description: "rewrite " + rel + " from the greeting setting", Safe: true, Files: []protocol.FileEdit{{Path: rel, Content: want}}}
		}
		res.Findings = append(res.Findings, finding)
	}
	return res, nil
}
