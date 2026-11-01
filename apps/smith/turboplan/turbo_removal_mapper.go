package turboplan

import (
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/turboworkspace"
)

var (
	managerCall = regexp.MustCompile(`(?m)(^|[\s;&|"'` + "`" + `(])(pnpm|yarn|turbo|corepack|TURBO_TOKEN|TURBO_TEAM)\b`)
	eslintTurbo = regexp.MustCompile(`eslint-(plugin|config)-turbo|turbo/no-undeclared-env-vars`)
)

// TurboPackage reports whether a package belongs to Turborepo, so a converted
// workspace no longer needs it.
func TurboPackage(name string) bool {
	return name == "turbo" || name == "turbo-ignore" || strings.HasPrefix(name, "@turbo/")
}

// planTurboRemoval schedules every Turborepo file and dependency for removal
// and reports the files that call turbo, pnpm or yarn.
func planTurboRemoval(plan *convertplan.Plan, w *turboworkspace.Workspace, files []string, read func(string) string) {
	plan.DeletePath(w.Turbo.File)
	for _, p := range w.Packages {
		if p.Turbo != nil {
			plan.DeletePath(p.Turbo.File)
		}
	}
	if hasPrefix(files, ".turbo/") {
		plan.DeletePath(".turbo")
	}
	plan.AddEdit("", w.Turbo.File, "removed: its tasks are the package.json scripts, with their dependencies, inputs, outputs and cache settings in smith.toml")

	var remove []string
	for _, deps := range []map[string]string{w.RootPackage.DevDependencies, w.RootPackage.Dependencies} {
		for name := range deps {
			if TurboPackage(name) {
				remove = append(remove, name)
			}
		}
	}
	sort.Strings(remove)
	plan.RemoveDependencies = remove
	if len(remove) > 0 {
		plan.AddEdit("", "package.json", "removes "+strings.Join(remove, ", "))
	}

	planPackageManager(plan, w, files, read)

	for _, f := range files {
		base := path.Base(f)
		switch {
		case strings.HasPrefix(base, "eslint.config.") || strings.HasPrefix(base, ".eslintrc"):
			if !strings.Contains(f, "node_modules/") && eslintTurbo.MatchString(read(f)) {
				plan.Add(convertplan.Manual, "", f, "uses eslint-plugin-turbo or eslint-config-turbo, which read turbo.json; remove the turbo rules and the two packages")
			}
		case isWorkflow(f) && f != ".github/workflows/ci.yml" && f != ".github/workflows/ci.yaml":
			if called := calls(read(f)); called != "" {
				plan.Add(convertplan.Manual, "", f, "calls "+called+"; run smith install and smith run instead (smith ci generate writes the pipeline), and drop the Turborepo remote cache variables")
			}
		case strings.HasPrefix(base, "Dockerfile"):
			if called := calls(read(f)); called != "" {
				plan.Add(convertplan.Manual, "", f, "calls "+called+"; for a Node app, smith run prune -p <app> writes the deploy output (production package.json and lockfile) that turbo prune --docker did, and the image installs it with npm ci --omit=dev")
			}
		}
	}
	if hasPrefix(files, ".changeset/") {
		plan.Add(convertplan.Changed, "", ".changeset", "smith releases from conventional commits (smith release), not changesets; the directory is kept, but nothing reads it any more")
	}
}

func isWorkflow(f string) bool {
	return strings.HasPrefix(f, ".github/workflows/") && (strings.HasSuffix(f, ".yml") || strings.HasSuffix(f, ".yaml"))
}

// calls names the tools a file mentions as commands.
func calls(text string) string {
	var found []string
	seen := map[string]bool{}
	for _, m := range managerCall.FindAllStringSubmatch(text, -1) {
		if !seen[m[2]] {
			seen[m[2]] = true
			found = append(found, m[2])
		}
	}
	sort.Strings(found)
	return strings.Join(found, ", ")
}
