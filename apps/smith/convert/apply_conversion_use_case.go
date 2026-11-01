package convert

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/lockfilesync"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// ApplyOptions configure Apply.
type ApplyOptions struct {
	Version string
	// SkipInstall leaves the lockfile alone (npm install is not run).
	SkipInstall bool
	// KeepPipelines leaves the CI pipelines as they are instead of
	// regenerating them (the full level's --ci keep).
	KeepPipelines bool
	Out           io.Writer
}

// Result is what Apply changed.
type Result struct {
	Pipelines []string `json:"pipelines,omitempty"`
	// Created are the files the conversion added that did not exist before
	// (smith.toml included); Rewritten are files it replaced.
	Created   []string `json:"created,omitempty"`
	Rewritten []string `json:"rewritten,omitempty"`
	// LostSlots are pipeline slots with steps that have no place in the new
	// pipeline; their steps must be moved by hand.
	LostSlots []string `json:"lostSlots,omitempty"`
}

// Apply carries out a plan: it writes smith.toml and the manifests, removes
// Nx, regenerates the pipelines and refreshes the lockfile. When a step before
// the lockfile fails (the converted workspace does not load, a file cannot be
// written), the files it touched are put back as they were and the error says
// so; a failed lockfile refresh leaves the conversion in place, since it is
// complete.
func Apply(plan *convertplan.Plan, opts ApplyOptions) (Result, error) {
	var res Result
	snap := newSnapshot(plan.Root)
	if err := guarded(func() error { return writeConversion(plan, opts, snap, &res) }); err != nil {
		if rerr := snap.restore(); rerr != nil {
			return Result{}, fmt.Errorf("%w; some files could not be put back (%v), so git restore . and git clean -fd return the tree to what it was", err, rerr)
		}
		return Result{}, fmt.Errorf("%w; nothing was changed: the files the conversion touched are back as they were", err)
	}
	if !opts.SkipInstall {
		fmt.Fprintf(opts.Out, "… reinstalling dependencies (this can take a minute)\n")
		notes, err := refreshLockfile(plan.Root, opts.Out, plan.Source != convertplan.SourceTurborepo)
		if err != nil {
			return res, err
		}
		fmt.Fprintf(opts.Out, "✔ reinstalled dependencies and updated package-lock.json\n")
		for _, n := range notes {
			fmt.Fprintln(opts.Out, "  "+n)
		}
	}
	return res, nil
}

// guarded runs step and turns a panic in it into an error, so Apply can put the
// files back instead of leaving the tree half converted.
func guarded(step func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("the conversion stopped on something it did not expect (%v)", r)
		}
	}()
	return step()
}

// writeConversion writes everything of a plan except the lockfile, remembering
// each file in snap before it changes it.
func writeConversion(plan *convertplan.Plan, opts ApplyOptions, snap *snapshot, res *Result) error {
	root := plan.Root
	file := func(rel string) string { return filepath.Join(root, filepath.FromSlash(rel)) }
	say := func(format string, args ...any) { fmt.Fprintf(opts.Out, format+"\n", args...) }
	for _, rel := range []string{workspace.FileName, "package.json", ".gitignore", "package-lock.json"} {
		snap.remember(rel)
	}
	for _, p := range plan.Projects {
		for _, manifest := range []string{"package.json", "pyproject.toml", "go.mod"} {
			snap.remember(p.Root + "/" + manifest)
		}
	}
	for _, rel := range plan.WriteFiles {
		snap.remember(rel)
	}
	for _, rel := range plan.Delete {
		snap.remember(rel)
	}
	for _, rel := range cipipeline.OwnedPaths {
		snap.remember(rel)
	}

	if err := os.WriteFile(file(workspace.FileName), []byte(ConfigText(plan.Config)), 0o644); err != nil {
		return err
	}
	res.Created = append(res.Created, workspace.FileName)
	say("✔ wrote %s", workspace.FileName)
	if err := editRootPackage(root, plan, opts.Version); err != nil {
		return err
	}
	manifests := 0
	for _, p := range plan.Projects {
		if !p.Supported {
			continue
		}
		if err := editManifest(root, p, plan.Specifiers); err != nil {
			return fmt.Errorf("%s: %w", p.Root, err)
		}
		manifests++
	}
	say("✔ updated the root package.json and %s", plural(manifests, "project manifest", "project manifests"))
	for _, rel := range plan.WriteFiles {
		if _, err := os.Stat(file(rel)); err == nil {
			res.Rewritten = append(res.Rewritten, rel)
		} else {
			res.Created = append(res.Created, rel)
		}
		if err := os.WriteFile(file(rel), []byte(plan.Write[rel]), 0o644); err != nil {
			return err
		}
	}
	if len(plan.WriteFiles) > 0 {
		say("✔ wrote %s", strings.Join(plan.WriteFiles, ", "))
	}
	for _, rel := range plan.Delete {
		if err := os.RemoveAll(file(rel)); err != nil {
			return err
		}
	}
	if len(plan.Delete) > 0 {
		say("✔ removed %s", plural(len(plan.Delete), "file or folder the old tool needed", "files and folders the old tool needed"))
	}
	if err := ignoreSmithFiles(root, plan); err != nil {
		return err
	}

	if opts.KeepPipelines {
		say("✔ kept the CI pipelines as they are")
	} else if err := regeneratePipelines(root, plan, opts, res); err != nil {
		return err
	}
	return nil
}

// regeneratePipelines renames the pipeline slots to smith's and regenerates
// the pipelines for the converted workspace.
func regeneratePipelines(root string, plan *convertplan.Plan, opts ApplyOptions, res *Result) error {
	file := func(rel string) string { return filepath.Join(root, filepath.FromSlash(rel)) }
	say := func(format string, args ...any) { fmt.Fprintf(opts.Out, format+"\n", args...) }
	for _, rel := range []string{cipipeline.GitHubPath, cipipeline.AzurePath} {
		raw, err := os.ReadFile(file(rel))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		renamed, orphans := RenameSlots(string(raw))
		for _, o := range orphans {
			res.LostSlots = append(res.LostSlots, rel+": "+o)
		}
		if err := os.WriteFile(file(rel), []byte(renamed), 0o644); err != nil {
			return err
		}
	}
	// Projects that import each other in a cycle (devDependencies may, in npm,
	// pnpm and Nx) load: doctor reports the cycle, it does not stop the conversion.
	s, err := session.Open(session.Options{Dir: root, Exact: true, Version: opts.Version, Log: opts.Out, AllowCycles: true})
	if err != nil {
		return fmt.Errorf("the converted workspace does not load: %w", err)
	}
	changed, lost, err := cipipeline.Regenerate(s)
	s.Close()
	if err != nil {
		return err
	}
	res.Pipelines = changed
	res.LostSlots = append(res.LostSlots, lost...)
	if len(changed) > 0 {
		say("✔ generated %s", strings.Join(changed, ", "))
	}
	res.Created = append(res.Created, changed...)
	return nil
}

// SmithPackage is the npm package that installs smith.
const SmithPackage = "monosmith"

// releasedVersion tells a smith release from a development build.
var releasedVersion = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// refreshLockfile drops what convert removed from package-lock.json (the Nx
// packages, unless pruneNx is false), then has lockfilesync install and settle
// it.
func refreshLockfile(root string, out io.Writer, pruneNx bool) ([]string, error) {
	lock := filepath.Join(root, "package-lock.json")
	if raw, err := os.ReadFile(lock); err == nil && pruneNx {
		pruned, dropped, err := PruneLockfile(raw, convertplan.NxPackage)
		if err != nil {
			return nil, err
		}
		if len(dropped) > 0 {
			if err := os.WriteFile(lock, pruned, 0o644); err != nil {
				return nil, err
			}
		}
	}
	res, err := lockfilesync.Sync(lockfilesync.Options{Root: root, Out: out, Create: true})
	if err != nil {
		return nil, fmt.Errorf("%w (the conversion itself is complete)", err)
	}
	return res.Notes, nil
}

// editRootPackage removes Nx from the root package.json and rewrites its
// scripts.
func editRootPackage(root string, plan *convertplan.Plan, version string) error {
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
			if deps.Len() == 0 && len(plan.RemoveDependencies) > 0 {
				doc.Delete(section)
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
		// A range the plan names (the package that replaces Nx's Jest plugins) is
		// used as it is; for the rest, Nx's install is the best guide.
		if rng := plan.AddDevDependencies[name]; rng != "" {
			deps.SetString(name, rng)
			continue
		}
		deps.SetString(name, installedRange(root, name))
	}
	// The root scripts call smith: a teammate who runs npm install and then
	// npm run build gets the same smith as everyone else.
	if releasedVersion.MatchString(version) {
		deps := doc.Object("devDependencies", true)
		if dependencies := doc.Object("dependencies", false); !deps.Has(SmithPackage) && (dependencies == nil || !dependencies.Has(SmithPackage)) {
			deps.SetString(SmithPackage, version)
		}
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
