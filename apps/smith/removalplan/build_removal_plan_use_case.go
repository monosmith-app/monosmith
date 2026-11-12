package removalplan

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

type groupBuilder struct {
	id    string
	build func(v *legacyinventory.Tree) built
}

// built is what one group's builder found.
type built struct {
	group *Group
	kept  []legacyinventory.Kept
	notes []string
}

// Build plans the removals the choices ask for, on the workspace t. Each
// group is computed on the tree the groups before it leave, so a package is
// only removed when nothing that stays uses it; a group in declined is left
// out and the groups after it are computed without it.
func Build(t *legacyinventory.Tree, c convertchoices.Choices, declined map[string]bool) Plan {
	plan := Plan{Choices: c}
	if c.Level != convertchoices.LevelFull {
		return plan
	}
	usage := legacyinventory.Detect(t)
	if c.CI == convertchoices.Keep {
		plan.Notes = append(plan.Notes, "--ci keep: your pipeline is not regenerated, so it still calls the old tool, and doctor reports \"CI pipelines match the workspace\" as failed until you run smith ci generate")
	}
	builders := []groupBuilder{}
	if c.Build == convertchoices.Smith {
		builders = append(builders, groupBuilder{GroupOutput, buildOutputGroup})
	}
	if c.Lint == convertchoices.Smith {
		builders = append(builders, groupBuilder{GroupLint, buildLintGroup})
	}
	switch c.Test {
	case convertchoices.Jest:
		builders = append(builders, groupBuilder{GroupTests, familyGroup(GroupTests, legacyinventory.FamilyVitest, testsReason("Vitest", "Jest"))})
		if !usage.Jest {
			plan.Notes = append(plan.Notes, "--test jest: no Jest config or package here; the tests are not rewritten to another runner, so nothing of the current one is removed")
		}
	case convertchoices.Vitest:
		builders = append(builders, groupBuilder{GroupTests, familyGroup(GroupTests, legacyinventory.FamilyJest, testsReason("Jest", "Vitest"))})
		if !usage.Vitest {
			plan.Notes = append(plan.Notes, "--test vitest: no Vitest config or package here; the tests are not rewritten to another runner, so nothing of the current one is removed")
		}
	}
	if c.Build == convertchoices.Smith {
		builders = append(builders, groupBuilder{GroupBuild, familyGroup(GroupBuild, legacyinventory.FamilyBuild, buildReason)})
	}
	builders = append(builders, groupBuilder{GroupNx, familyGroup(GroupNx, legacyinventory.FamilyNx, nxReason)})

	v := t.Clone()
	keptSeen := map[string]bool{}
	for _, b := range builders {
		if declined[b.id] {
			continue
		}
		out := b.build(v)
		for _, k := range out.kept {
			if !keptSeen[k.What] {
				keptSeen[k.What] = true
				plan.Kept = append(plan.Kept, k)
			}
		}
		plan.Notes = append(plan.Notes, out.notes...)
		if out.group != nil {
			v.Apply(out.group.Fix.Fix)
			plan.Groups = append(plan.Groups, *out.group)
		}
	}
	if left := legacyinventory.Mentioning(v, 6, "nx", "@nx", "mnci", "@mnci"); len(left) > 0 {
		plan.Notes = append(plan.Notes, "still mentions Nx or mnci, which nothing here removes: "+strings.Join(left, ", ")+" (check them by hand)")
	}
	return plan
}

func testsReason(removed, chosen string) func(files, pkgs int) string {
	return func(files, pkgs int) string {
		return fmt.Sprintf("Remove %s: you chose %s", counts(pkgs, "package", files, removed+" config file"), chosen)
	}
}

func buildReason(files, pkgs int) string {
	return fmt.Sprintf("Remove %s of the old build: smith builds this stack", counts(pkgs, "build package", files, "build config file"))
}

func nxReason(files, pkgs int) string {
	return fmt.Sprintf("Remove %s: nothing runs Nx or mnci any more", counts(pkgs, "Nx or mnci package", files, "Nx or mnci file"))
}

// counts spells "4 lint packages and 3 config files"; a zero is left out.
func counts(pkgs int, pkgNoun string, files int, fileNoun string) string {
	var parts []string
	if pkgs > 0 {
		parts = append(parts, plural(pkgs, pkgNoun))
	}
	if files > 0 {
		parts = append(parts, plural(files, fileNoun))
	}
	return strings.Join(parts, " and ")
}

func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// familyGroup builds the group that removes one family's files, then the
// packages nothing needs once those files are gone.
func familyGroup(id string, family legacyinventory.Family, reason func(files, pkgs int) string) func(v *legacyinventory.Tree) built {
	return func(v *legacyinventory.Tree) built {
		files := legacyinventory.ScanFiles(v, []legacyinventory.Family{family})
		out := built{kept: files.Kept}
		var edits []protocol.FileEdit
		work := v.Clone()
		for _, f := range files.Files {
			edits = append(edits, protocol.FileEdit{Path: f.Path, Delete: true})
			work.Delete(f.Path)
		}
		pkgs := legacyinventory.ScanPackages(work, []legacyinventory.Family{family})
		out.kept = append(out.kept, pkgs.Kept...)
		manifestEdits, names, err := removePackageEdits(work, pkgs.Packages, nil)
		if err != nil {
			out.notes = append(out.notes, err.Error())
			return out
		}
		edits = append(edits, manifestEdits...)
		if len(edits) == 0 {
			return out
		}
		out.group = newGroup(id, reason(len(files.Files), len(names)), len(files.Files), names, edits)
		return out
	}
}

// buildLintGroup moves the root ESLint config over to smith's, removes the
// other lint configs, and the lint packages nothing needs afterwards.
func buildLintGroup(v *legacyinventory.Tree) built {
	var out built
	cfg := legacyinventory.RootLintConfig(v)
	files := legacyinventory.ScanFiles(v, []legacyinventory.Family{legacyinventory.FamilyLint})
	out.kept = files.Kept
	work := v.Clone()
	var edits []protocol.FileEdit
	rewritten := ""
	if cfg != "" {
		text := v.Read(cfg)
		next, how, why, ok := SwapLintConfig(text)
		if !ok {
			out.notes = append(out.notes, fmt.Sprintf("--lint smith: %s stays as it is: %s", cfg, why))
			out.kept = append(out.kept, legacyinventory.Kept{What: cfg, Family: legacyinventory.FamilyLint, Why: why})
			return out
		}
		if next != text {
			edits = append(edits, protocol.FileEdit{Path: cfg, Content: next})
			work.Write(cfg, next)
			rewritten = how
		}
	}
	removedFiles := 0
	for _, f := range files.Files {
		if f.Path == cfg {
			continue
		}
		edits = append(edits, protocol.FileEdit{Path: f.Path, Delete: true})
		work.Delete(f.Path)
		removedFiles++
	}
	if cfg == "" && removedFiles == 0 {
		return out
	}
	pkgs := legacyinventory.ScanPackages(work, []legacyinventory.Family{legacyinventory.FamilyLint})
	out.kept = append(out.kept, pkgs.Kept...)
	// smith's config needs a newer ESLint than many workspaces pin: every
	// manifest that names eslint is raised with it, or the plan says why not.
	var raisedFrom []string
	manifestEdits, names, err := removePackageEdits(work, pkgs.Packages, func(path, text string) (string, error) {
		next, raised, problem, err := RaiseESLint(text)
		if err != nil {
			return "", err
		}
		raisedFrom = append(raisedFrom, raised...)
		if problem != "" {
			out.notes = append(out.notes, fmt.Sprintf("--lint smith: %s: %s. Set eslint to %s there before you convert, or ESLint cannot load smith's config and the lint check reports a crash", path, problem, LintESLintRange))
		}
		if path != "package.json" {
			return next, nil
		}
		return AddDevDependency(next, LintConfigPackage, LintConfigRange)
	})
	if err != nil {
		out.notes = append(out.notes, err.Error())
		return out
	}
	edits = append(edits, manifestEdits...)
	if len(edits) == 0 {
		return out
	}
	reason := "Use smith's ESLint config (" + LintConfigPackage + ")"
	if c := counts(len(names), "old lint package", removedFiles, "other lint config file"); c != "" {
		reason += ", remove " + c
	}
	reason += ": you chose smith's ESLint"
	if len(raisedFrom) > 0 {
		reason += "; raises eslint " + majorOf(raisedFrom[0]) + " → " + majorOf(LintESLintFloor) + ", which smith's config needs"
	}
	switch rewritten {
	case LintSwapped:
		reason += "; the rules you wrote in " + cfg + " stay"
	case LintReplaced:
		reason += "; " + cfg + " is replaced by smith's two lines, so the rules written in it are gone (the diff shows them)"
	}
	out.group = newGroup(GroupLint, reason, removedFiles+boolCount(rewritten != ""), names, edits)
	return out
}

// majorOf is a range's major version ("^9.39.3" is "9"), or the range itself
// when it has no plain lowest version.
func majorOf(r string) string {
	floor, ok := rangeFloor(r)
	if !ok {
		return r
	}
	major, _, _ := strings.Cut(floor, ".")
	return major
}

func boolCount(b bool) int {
	if b {
		return 1
	}
	return 0
}

// removePackageEdits edits each package.json that declares a package being
// removed. extra may change a manifest further: it is called for every
// package.json of the workspace, also those nothing is removed from.
func removePackageEdits(work *legacyinventory.Tree, pkgs []legacyinventory.LegacyPackage, extra func(path, text string) (string, error)) ([]protocol.FileEdit, []string, error) {
	byManifest := map[string][]string{}
	var all []string
	seen := map[string]bool{}
	for _, p := range pkgs {
		for _, mf := range p.In {
			byManifest[mf] = append(byManifest[mf], p.Name)
		}
		if !seen[p.Name] {
			seen[p.Name] = true
			all = append(all, p.Name)
		}
	}
	sort.Strings(all)
	if extra != nil {
		for _, mf := range work.ManifestPaths() {
			if _, ok := byManifest[mf]; !ok {
				byManifest[mf] = nil
			}
		}
	}
	var paths []string
	for mf := range byManifest {
		paths = append(paths, mf)
	}
	sort.Strings(paths)
	var edits []protocol.FileEdit
	for _, mf := range paths {
		text := work.Read(mf)
		next, err := RemovePackages(text, byManifest[mf])
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", mf, err)
		}
		if extra != nil {
			if next, err = extra(mf, next); err != nil {
				return nil, nil, fmt.Errorf("%s: %w", mf, err)
			}
		}
		if next != text {
			edits = append(edits, protocol.FileEdit{Path: mf, Content: next})
		}
	}
	return edits, all, nil
}

func newGroup(id, reason string, files int, packages []string, edits []protocol.FileEdit) *Group {
	fix := protocol.Fix{Description: reason, Safe: false, Files: edits}
	return &Group{
		ID: id, Reason: reason, Files: files, Packages: len(packages), Removes: packages,
		Fix: doctorfix.Planned{Fix: fix, Checks: []string{"convert --level full: " + id}, Ask: true},
	}
}
