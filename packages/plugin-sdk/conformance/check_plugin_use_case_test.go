package conformance

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// good is a language whose manifest, fig.txt, holds "kind stack".
type good struct{ root string }

func (g *good) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	g.root = p.WorkspaceRoot
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "fig", Version: "1.0.0"}, Kind: protocol.KindLanguage, Manifests: []string{"fig.txt"},
		Stacks: []protocol.Stack{{Name: "fig-plain", Language: "fig"}}}, nil
}

func (g *good) Generate(p protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(g.root, filepath.FromSlash(p.Directory))
	_ = os.MkdirAll(dir, 0o755)
	return protocol.GenerateResult{Files: []string{p.Directory + "/fig.txt"}}, os.WriteFile(filepath.Join(dir, "fig.txt"), []byte(string(p.Kind)+" "+p.Stack), 0o644)
}

func (g *good) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	var res protocol.DiscoverResult
	for _, m := range p.Manifests {
		raw, err := os.ReadFile(filepath.Join(g.root, filepath.FromSlash(m)))
		if err != nil {
			return res, err
		}
		kind, stack, _ := strings.Cut(string(raw), " ")
		res.Projects = append(res.Projects, protocol.Project{Name: path.Base(path.Dir(m)), Root: path.Dir(m), Kind: protocol.ProjectKind(kind), Stack: stack,
			Tasks: map[string]protocol.Task{"build": {Command: []string{"fig", "build"}}}})
	}
	return res, nil
}

// bad breaks several promises: a structure plugin that does nothing,
// providing a task, publishing an incomplete stack.
type bad struct{}

func (bad) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "bad"}, Kind: protocol.KindConvention, Provides: []string{"test"},
		Stacks: []protocol.Stack{{Name: "half"}}}, nil
}

func connect(t *testing.T, p serve.Plugin, ws string) *host.Client {
	t.Helper()
	toPlugin, fromHost := io.Pipe()
	toHost, fromPlugin := io.Pipe()
	go func() { _ = serve.Serve(p, toPlugin, fromPlugin) }()
	c, err := host.Connect(toHost, fromHost, host.Options{WorkspaceRoot: ws})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestAGoodPluginConforms(t *testing.T) {
	ws := t.TempDir()
	r := Check(connect(t, &good{}, ws), Options{Workspace: ws})
	if !r.OK() {
		t.Fatalf("failures:\n%s", r.Failures())
	}
	var checks []string
	for _, res := range r.Results {
		checks = append(checks, res.Check)
	}
	for _, want := range []string{"discovers the generated package", "the generated app records its stack", "stack fig-plain is complete"} {
		if !strings.Contains(strings.Join(checks, "|"), want) {
			t.Errorf("check %q did not run: %v", want, checks)
		}
	}
}

func TestABadPluginIsReported(t *testing.T) {
	r := Check(connect(t, bad{}, t.TempDir()), Options{})
	failures := r.Failures()
	for _, want := range []string{"names itself", "a structure plugin does something", "only a toolchain provides tasks", "stack half is complete"} {
		if !strings.Contains(failures, want) {
			t.Errorf("not reported: %s\n%s", want, failures)
		}
	}
}

// badRefs is a toolchain that fills no slot and declares references that
// name nothing, an unknown level, and itself.
type badRefs struct{}

func (badRefs) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{
		Plugin: protocol.Peer{Name: "refs", Version: "1.0.0"}, Kind: protocol.KindConvention,
		Roles:     []protocol.Role{protocol.RoleToolchain},
		Requires:  []protocol.PluginRef{{}, {Level: "platform", Name: "x"}},
		Conflicts: []protocol.PluginRef{protocol.ConflictsWith("refs")},
	}, nil
}

func (badRefs) Contribute(protocol.ContributeParams) (protocol.ContributeResult, error) {
	return protocol.ContributeResult{}, nil
}

func TestLevelsAndReferencesAreChecked(t *testing.T) {
	failures := Check(connect(t, badRefs{}, t.TempDir()), Options{}).Failures()
	for _, want := range []string{"a toolchain names the slots it fills", "requires plugin: a known reference", "requires platform plugin x: a known reference", "conflicts with plugin refs: a known reference"} {
		if !strings.Contains(failures, want) {
			t.Errorf("not reported: %s\n%s", want, failures)
		}
	}
}

// scoped is a convention with the given languages.
type scoped struct{ languages []string }

func (s scoped) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "scoped", Version: "1.0.0"}, Kind: protocol.KindConvention, Languages: s.languages}, nil
}

func (scoped) Contribute(protocol.ContributeParams) (protocol.ContributeResult, error) {
	return protocol.ContributeResult{}, nil
}

func checkNames(r Report) string {
	var names []string
	for _, res := range r.Results {
		names = append(names, res.Check)
	}
	return strings.Join(names, "|")
}

func TestAConventionSaysWhichLanguagesItAppliesTo(t *testing.T) {
	r := Check(connect(t, scoped{languages: []string{"go", "python"}}, t.TempDir()), Options{})
	if !strings.Contains(checkNames(r), "declares the languages it applies to: go, python") {
		t.Fatalf("languages not reported: %s", checkNames(r))
	}
}

func TestAConventionWithoutLanguagesSaysItMeansEveryLanguage(t *testing.T) {
	r := Check(connect(t, scoped{}, t.TempDir()), Options{})
	if !strings.Contains(checkNames(r), "declares no languages, so it applies to projects of every language") || !r.OK() {
		t.Fatalf("%s\n%s", checkNames(r), r.Failures())
	}
}

func TestBlankOrRepeatedLanguagesAreReported(t *testing.T) {
	r := Check(connect(t, scoped{languages: []string{"go", "go", " "}}, t.TempDir()), Options{})
	if !strings.Contains(r.Failures(), "declares the languages it applies to") {
		t.Fatalf("not reported:\n%s", r.Failures())
	}
}
