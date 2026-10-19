package convert

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// ApplyOptions configure Apply.
type ApplyOptions struct {
	Version string
	// SkipInstall leaves the lockfile alone (npm install is not run).
	SkipInstall bool
	Out         io.Writer
}

// Result is what Apply changed.
type Result struct {
	Pipelines []string `json:"pipelines,omitempty"`
	// LostSlots are pipeline slots with steps that have no place in the new
	// pipeline; their steps must be moved by hand.
	LostSlots []string `json:"lostSlots,omitempty"`
}

// Apply carries out a plan: it writes smith.toml and the manifests, removes
// Nx, regenerates the pipelines and refreshes the lockfile.
func Apply(plan *convertplan.Plan, opts ApplyOptions) (Result, error) {
	var res Result
	root := plan.Root
	file := func(rel string) string { return filepath.Join(root, filepath.FromSlash(rel)) }
	say := func(format string, args ...any) { fmt.Fprintf(opts.Out, format+"\n", args...) }

	if err := os.WriteFile(file(workspace.FileName), []byte(ConfigText(plan.Config)), 0o644); err != nil {
		return res, err
	}
	say("wrote %s", workspace.FileName)
	if err := editRootPackage(root, plan); err != nil {
		return res, err
	}
	for _, p := range plan.Projects {
		if !p.Supported {
			continue
		}
		if err := editManifest(root, p, plan.Specifiers); err != nil {
			return res, fmt.Errorf("%s: %w", p.Root, err)
		}
	}
	for _, rel := range plan.WriteFiles {
		if err := os.WriteFile(file(rel), []byte(plan.Write[rel]), 0o644); err != nil {
			return res, err
		}
		say("wrote %s", rel)
	}
	for _, rel := range plan.Delete {
		if err := os.RemoveAll(file(rel)); err != nil {
			return res, err
		}
		say("removed %s", rel)
	}
	if err := ignoreSmithFiles(root, plan); err != nil {
		return res, err
	}

	for _, rel := range []string{cipipeline.GitHubPath, cipipeline.AzurePath} {
		raw, err := os.ReadFile(file(rel))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return res, err
		}
		renamed, orphans := RenameSlots(string(raw))
		for _, o := range orphans {
			res.LostSlots = append(res.LostSlots, rel+": "+o)
		}
		if err := os.WriteFile(file(rel), []byte(renamed), 0o644); err != nil {
			return res, err
		}
	}
	s, err := session.Open(session.Options{Dir: root, Version: opts.Version, Log: opts.Out})
	if err != nil {
		return res, fmt.Errorf("the converted workspace does not load: %w", err)
	}
	changed, lost, err := cipipeline.Regenerate(s)
	s.Close()
	if err != nil {
		return res, err
	}
	res.Pipelines = changed
	res.LostSlots = append(res.LostSlots, lost...)
	for _, c := range changed {
		say("generated %s", c)
	}

	if !opts.SkipInstall {
		if err := refreshLockfile(root, opts.Out, plan.Source != convertplan.SourceTurborepo); err != nil {
			return res, err
		}
	}
	return res, nil
}

// refreshLockfile runs npm install so package-lock.json drops what convert
// removed, and checks that npm ci accepts the result. npm 11 can reuse the
// installed tree on the first run and leave the lockfile out of step, so a
// second install re-resolves it when the check fails.
func refreshLockfile(root string, out io.Writer, pruneNx bool) error {
	if _, err := os.Stat(filepath.Join(root, "package.json")); err != nil {
		return nil
	}
	if _, err := exec.LookPath("npm"); err != nil {
		fmt.Fprintln(out, "npm is not installed: run npm install to update package-lock.json")
		return nil
	}
	lock := filepath.Join(root, "package-lock.json")
	if raw, err := os.ReadFile(lock); err == nil && pruneNx {
		pruned, dropped, err := PruneLockfile(raw, convertplan.NxPackage)
		if err != nil {
			return err
		}
		if len(dropped) > 0 {
			if err := os.WriteFile(lock, pruned, 0o644); err != nil {
				return err
			}
			fmt.Fprintf(out, "removed %d Nx packages from package-lock.json\n", len(dropped))
		}
	}
	run := func(args ...string) error {
		cmd := exec.Command("npm", args...)
		cmd.Dir, cmd.Stdout, cmd.Stderr = root, out, out
		return cmd.Run()
	}
	quiet := func(args ...string) error {
		cmd := exec.Command("npm", args...)
		cmd.Dir = root
		return cmd.Run()
	}
	for attempt := 1; attempt <= 2; attempt++ {
		fmt.Fprintln(out, "running npm install to update the lockfile")
		if err := run("install", "--no-audit", "--no-fund"); err != nil {
			return fmt.Errorf("npm install failed (%w); fix it and run npm install again, the conversion itself is complete", err)
		}
		if quiet("ci", "--dry-run", "--ignore-scripts", "--no-audit", "--no-fund") == nil {
			return nil
		}
	}
	return errors.New("package-lock.json is still out of step with package.json after npm install; run npm install until `npm ci --dry-run` passes")
}

// editRootPackage removes Nx from the root package.json and rewrites its
// scripts.
func editRootPackage(root string, plan *convertplan.Plan) error {
	path := filepath.Join(root, "package.json")
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return fmt.Errorf("package.json: %w", err)
	}
	doc.Delete("nx")
	for _, field := range plan.RemoveFields {
		deleteField(doc, field)
	}
	if plan.SetWorkspaces != nil {
		if err := doc.SetJSON("workspaces", plan.SetWorkspaces); err != nil {
			return err
		}
	}
	if len(plan.SetOverrides) > 0 {
		overrides := doc.Object("overrides", true)
		for _, name := range sortedKeys(plan.SetOverrides) {
			overrides.SetString(name, plan.SetOverrides[name])
		}
	}
	applySpecifiers(doc, plan.Specifiers["."])
	for _, section := range []string{"devDependencies", "dependencies"} {
		if deps := doc.Object(section, false); deps != nil {
			for _, name := range plan.RemoveDependencies {
				deps.Delete(name)
			}
		}
	}
	if overrides := doc.Object("overrides", false); overrides != nil {
		for _, name := range plan.RemoveOverrides {
			overrides.Delete(name)
		}
		if overrides.Len() == 0 {
			doc.Delete("overrides")
		}
	}
	for _, name := range sortedKeys(plan.AddDevDependencies) {
		deps := doc.Object("devDependencies", true)
		if doc.Object("dependencies", false) != nil && doc.Object("dependencies", false).Has(name) || deps.Has(name) {
			continue
		}
		deps.SetString(name, installedRange(root, name))
	}
	scripts := doc.Object("scripts", len(plan.RootScripts) > 0)
	for _, name := range sortedKeys(plan.RootScripts) {
		if cmd := plan.RootScripts[name]; cmd == "" {
			scripts.Delete(name)
		} else {
			scripts.SetString(name, cmd)
		}
	}
	return os.WriteFile(path, doc.Bytes(), 0o644)
}

// deleteField removes a root package.json field; "engines.pnpm" is the pnpm
// key of engines, which goes too when it leaves engines empty.
func deleteField(doc *jsondoc.Document, field string) {
	parent, key, nested := strings.Cut(field, ".")
	if !nested {
		doc.Delete(parent)
		return
	}
	obj := doc.Object(parent, false)
	if obj == nil {
		return
	}
	obj.Delete(key)
	if obj.Len() == 0 {
		doc.Delete(parent)
	}
}

// applySpecifiers rewrites the range of each named dependency, in every
// dependency section that lists it.
func applySpecifiers(doc *jsondoc.Document, specifiers map[string]string) {
	if len(specifiers) == 0 {
		return
	}
	for _, section := range []string{"dependencies", "devDependencies", "peerDependencies", "optionalDependencies"} {
		deps := doc.Object(section, false)
		if deps == nil {
			continue
		}
		for _, name := range sortedKeys(specifiers) {
			if deps.Has(name) {
				deps.SetString(name, specifiers[name])
			}
		}
	}
}

// installedRange is ^<installed version> of a package already in
// node_modules (Nx installed it), or "*" when it is not there.
func installedRange(root, name string) string {
	raw, err := os.ReadFile(filepath.Join(root, "node_modules", filepath.FromSlash(name), "package.json"))
	if err != nil {
		return "*"
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(raw, &pkg) != nil || pkg.Version == "" {
		return "*"
	}
	return "^" + pkg.Version
}

// editManifest records scripts and the kind in a project's manifest, and
// removes the package.json "nx" field.
func editManifest(root string, p *convertplan.ProjectPlan, specifiers map[string]map[string]string) error {
	dir := filepath.Join(root, filepath.FromSlash(p.Root))
	switch p.Language {
	case "node":
		path := filepath.Join(dir, "package.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		doc, err := jsondoc.ParseDocument(raw)
		if err != nil {
			return err
		}
		doc.Delete("nx")
		applySpecifiers(doc, specifiers[p.Root])
		if len(p.Scripts) > 0 {
			scripts := doc.Object("scripts", true)
			for _, name := range sortedKeys(p.Scripts) {
				scripts.SetString(name, p.Scripts[name])
			}
		}
		if p.KindMarker {
			doc.Object("smith", true).SetString("kind", string(p.Kind))
		}
		return os.WriteFile(path, doc.Bytes(), 0o644)
	case "python":
		if !p.KindMarker {
			return nil
		}
		return appendIfMissing(filepath.Join(dir, "pyproject.toml"), "[tool.smith]", "\n[tool.smith]\nkind = \""+string(p.Kind)+"\"\n")
	case "go":
		if !p.KindMarker {
			return nil
		}
		return appendIfMissing(filepath.Join(dir, "go.mod"), "smith:kind", "\n// smith:kind "+string(p.Kind)+"\n")
	}
	return nil
}

func appendIfMissing(path, marker, text string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.Contains(string(raw), marker) {
		return nil
	}
	return os.WriteFile(path, append([]byte(strings.TrimRight(string(raw), "\n")+"\n"), text...), 0o644)
}

// ignoreSmithFiles adds smith's cache (and the Python venv) to .gitignore.
func ignoreSmithFiles(root string, plan *convertplan.Plan) error {
	path := filepath.Join(root, ".gitignore")
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	lines := map[string]bool{}
	for _, l := range strings.Split(string(raw), "\n") {
		lines[strings.TrimSpace(l)] = true
	}
	var add []string
	want := []string{".smith/"}
	for _, p := range plan.Projects {
		if p.Language == "python" {
			want = append(want, ".venv/")
			break
		}
	}
	for _, w := range want {
		if !lines[w] && !lines[strings.TrimSuffix(w, "/")] && !lines["/"+w] {
			add = append(add, w)
		}
	}
	if len(add) == 0 {
		return nil
	}
	text := strings.TrimRight(string(raw), "\n")
	if text != "" {
		text += "\n\n"
	}
	text += "# smith\n" + strings.Join(add, "\n") + "\n"
	return os.WriteFile(path, []byte(text), 0o644)
}
