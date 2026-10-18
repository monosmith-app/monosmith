package doctor

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/releaselineage"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Report is `smith doctor --json`.
type Report struct {
	Findings []Finding `json:"findings"`
	Passed   int       `json:"passed"`
	Failed   int       `json:"failed"`
	Warnings int       `json:"warnings"`
}

// Finding is one check; Source is "smith" or the plugin's name.
type Finding struct {
	protocol.Finding
	Source string `json:"source"`
}

// Run collects every finding.
func Run(s *session.Session) Report {
	var findings []Finding
	add := func(source string, fs ...protocol.Finding) {
		for _, f := range fs {
			findings = append(findings, Finding{Finding: f, Source: source})
		}
	}
	add("smith", CoreChecks(s)...)
	add("smith", lineageChecks(s)...)
	add("smith", recipeChecks(s)...)
	for _, c := range s.Plugins.Clients {
		if !c.Info.Capabilities.Doctor {
			continue
		}
		var projects []protocol.Project
		for _, p := range s.Graph.Projects {
			if p.ServedBy(c.Info.Plugin.Name) {
				projects = append(projects, p.Project)
			}
		}
		var res protocol.DoctorResult
		if err := c.Call(protocol.MethodDoctor, protocol.DoctorParams{Projects: projects}, &res, time.Minute); err != nil {
			add(c.Info.Plugin.Name, protocol.Finding{Check: "plugin " + c.Info.Plugin.Name + " answers doctor", Detail: err.Error(), Remedy: "report this to the plugin's maintainers"})
			continue
		}
		add(c.Info.Plugin.Name, res.Findings...)
	}
	report := Report{Findings: findings}
	for _, f := range findings {
		switch {
		case f.OK:
			report.Passed++
		case f.Warning:
			report.Warnings++
		default:
			report.Failed++
		}
	}
	sort.SliceStable(report.Findings, func(i, j int) bool { return !report.Findings[i].OK && report.Findings[j].OK })
	return report
}

// lineageChecks report release tags a release would refuse to start from:
// under an old name, or behind the registry (this asks each registry).
func lineageChecks(s *session.Session) []protocol.Finding {
	failed := map[string]bool{}
	var out []protocol.Finding
	for _, f := range release.CheckLineage(s, false) {
		failed[f.Check] = true
		out = append(out, protocol.Finding{Check: f.Check, Detail: f.Message, Warning: !f.Blocking, Remedy: "see the detail; smith release refuses to run until it is fixed"})
	}
	for _, check := range []string{releaselineage.CheckResolve, releaselineage.CheckRegistry} {
		if !failed[check] {
			out = append(out, protocol.Finding{Check: check, OK: true})
		}
	}
	return out
}

// CoreChecks are smith's own invariants.
func CoreChecks(s *session.Session) []protocol.Finding {
	var out []protocol.Finding
	_, err := s.Git.Run("check-ignore", ".smith/cache/x")
	ignoreCache := protocol.Finding{
		Check: "the cache directory is ignored by git", OK: err == nil,
		Detail: ".smith/ is not in .gitignore, so cache files could be committed", Remedy: "add `.smith/` to .gitignore",
	}
	if err != nil {
		ignoreCache.Fix = ignoreFix(s.Root)
	}
	out = append(out, ignoreCache)
	out = append(out, protocol.Finding{
		Check: "the repository has commits", OK: s.Git.HasCommits(),
		Detail: "releases and affected detection need history", Remedy: "commit the workspace",
	})
	var ignored, tracked []string
	if raw, err := s.Git.Run("ls-files", "--cached", "--ignored", "--exclude-standard"); err == nil {
		ignored = lines(raw)
	}
	if raw, err := s.Git.Run("ls-files"); err == nil {
		tracked = lines(raw)
	}
	generated := TrackedGenerated(ignored, tracked, s.Graph.Projects)
	out = append(out, protocol.Finding{
		Check: "no task rewrites a tracked file", OK: len(generated) == 0, Warning: true,
		Detail: "a run leaves these changed, and the release refuses a dirty tree (smith ci verify fails on them): " + strings.Join(generated, "; "),
		Remedy: "stop tracking generated files: `git rm --cached <file>` (keep or add its .gitignore rule) and commit; a file a task updates on purpose is committed with the change",
	})
	remote := s.Config.Cache.Remote
	if env := os.Getenv("SMITH_CACHE_REMOTE"); env != "" {
		remote = env
	}
	out = append(out, cacheKeyFinding(remote, os.Getenv)...)
	outside := OutsideBuildDir(s.Graph.Projects)
	out = append(out, protocol.Finding{
		Check: "every build writes to " + protocol.BuildDir + "/", OK: len(outside) == 0,
		Detail: strings.Join(outside, "; "),
		Remedy: "build each project into " + protocol.BuildDir + "/<project dir>/: point the tool there (tsconfig outDir, --outdir, -o) and the task's outputs (smith.toml outputs = [...]); for a plugin's own task, update the plugin",
	})
	cycles := s.Graph.Cycles()
	var paths []string
	for _, c := range cycles {
		paths = append(paths, strings.Join(c, " → "))
	}
	out = append(out, protocol.Finding{
		Check: "projects do not import each other in a cycle", OK: len(cycles) == 0,
		Detail: "import cycles: " + strings.Join(paths, "; ") + ". Tasks cannot be ordered and a release cannot version them",
		Remedy: "move what both sides need into a project of its own that neither imports back, or merge the projects",
	})
	out = append(out, protocol.Finding{
		Check: "the clone has full history", OK: !s.Git.IsShallow(), Warning: true,
		Detail: "a shallow clone cannot compute versions or affected projects", Remedy: "`git fetch --unshallow` (in CI: checkout with fetch-depth 0)",
	})
	if c, err := cipipeline.Collect(s); err == nil {
		model := cipipeline.NewModel(c, s.Version, s.Config.Release.Branch)
		drift := cipipeline.Drift(s.Root, s.Config, model)
		detail := ""
		if len(drift) > 0 {
			detail = drift[0]
		}
		pipelines := protocol.Finding{
			Check: "CI pipelines match the workspace", OK: len(drift) == 0,
			Detail: detail, Remedy: "run `smith ci generate` and commit the result",
		}
		planned, lost := cipipeline.Planned(s.Root, s.Config, model)
		stale := cipipeline.Stale(s.Root, cipipeline.Rendered(s.Config, model))
		if len(planned)+len(stale) > 0 {
			// Regenerating is safe when every slot's content survives it; a
			// stale file carries smith's header, so it is smith's to delete.
			pipelines.Fix = &protocol.Fix{Description: "regenerate the pipelines, keeping your slot contents", Safe: len(lost) == 0}
			for _, path := range sortedKeys(planned) {
				pipelines.Fix.Files = append(pipelines.Fix.Files, protocol.FileEdit{Path: path, Content: planned[path]})
			}
			for _, path := range stale {
				pipelines.Fix.Files = append(pipelines.Fix.Files, protocol.FileEdit{Path: path, Delete: true})
			}
		}
		out = append(out, pipelines)
	}
	for _, p := range s.Graph.Projects {
		missing := []string{}
		for _, t := range []string{"lint", "test"} {
			if _, ok := p.Tasks[t]; !ok {
				missing = append(missing, t)
			}
		}
		if len(missing) > 0 {
			out = append(out, protocol.Finding{
				Check: p.Name + " can be verified", Warning: true,
				Detail: p.Name + " has no " + missing[0] + " task, so CI never checks it", Remedy: "add the task to the project (or in smith.toml under [projects." + p.Name + ".tasks])",
			})
		}
		out = append(out, droppedDependencies(p)...)
	}
	return out
}

// droppedDependencies names the dependencies a plugin gave a task that a
// smith.toml depends_on override leaves out: an override replaces the list,
// so a dependency it does not repeat is gone without a word.
func droppedDependencies(p *projectgraph.Project) []protocol.Finding {
	var out []protocol.Finding
	for _, task := range slices.Sorted(maps.Keys(p.Dropped)) {
		// The source names the override, and any dependency it kept after it.
		override, _, _ := strings.Cut(p.Sources[task+".dependsOn"], ", with ")
		out = append(out, protocol.Finding{
			Check: p.Name + ":" + task + " keeps the dependencies its plugins add", Warning: true,
			Detail: "depends_on in " + override + " leaves out " + strings.Join(p.Dropped[task], ", "),
			Remedy: "add them to depends_on, or leave them out if the task does not need them (`smith explain " + p.Name + ":" + task + "` shows where each field comes from)",
		})
	}
	return out
}

// ignoreFix adds .smith/ to the workspace's .gitignore (creating it).
func ignoreFix(root string) *protocol.Fix {
	raw, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
	text := string(raw)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return &protocol.Fix{Description: "ignore the cache directory", Safe: true, Files: []protocol.FileEdit{{Path: ".gitignore", Content: text + "\n# smith\n.smith/\n"}}}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
