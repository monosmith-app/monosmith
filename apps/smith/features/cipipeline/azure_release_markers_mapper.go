package cipipeline

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

// buildNumberLimit is the longest build number Azure Pipelines accepts.
const buildNumberLimit = 255

// azureReleaseMarkers are the logging commands that tell Azure Pipelines what
// the release plan holds, as [ci.azure] asks: a build tag for each planned
// project, DEPLOY_<PROJECT> = true or false for each deployable project, and a
// build number made from the planned projects. No planned project leaves the
// build number alone.
//
// Parameters: cfg is [ci.azure]; plan is the release plan; deployables are the
// names of the apps and the projects with a package task, in a stable order.
// Returns: the ##vso lines, in that order.
func azureReleaseMarkers(cfg workspace.AzureSection, plan release.Plan, deployables []string) []string {
	var lines []string
	planned := map[string]bool{}
	for _, item := range plan.Items {
		planned[item.Project] = true
		if tag, on := cfg.BuildTag(item.Project); on {
			lines = append(lines, "##vso[build.addbuildtag]"+tag)
		}
	}
	if cfg.DeployVariablesOn() {
		for _, name := range deployables {
			value := "false"
			if planned[name] {
				value = "true"
			}
			lines = append(lines, "##vso[task.setvariable variable="+deployVariable(name)+"]"+value)
		}
	}
	if number := buildNumber(cfg.BuildNumberPattern(), plan); number != "" {
		lines = append(lines, "##vso[build.updatebuildnumber]"+number)
	}
	return lines
}

// deployVariable is the DEPLOY_<PROJECT> name of a project: upper case, with
// every character that is not a letter or a digit replaced by _.
func deployVariable(project string) string {
	var b strings.Builder
	b.WriteString("DEPLOY_")
	for _, r := range strings.ToUpper(project) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}

// buildNumber renders pattern for each planned project and joins the results
// with "-". The characters Azure refuses become _, and the number is cut to
// Azure's limit without a trailing dot. It is "" when the pattern is empty or
// the plan has no project.
func buildNumber(pattern string, plan release.Plan) string {
	if pattern == "" || len(plan.Items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(plan.Items))
	for _, item := range plan.Items {
		parts = append(parts, strings.NewReplacer("{project}", item.Project, "{version}", item.Next).Replace(pattern))
	}
	number := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`"/:<>\|?@*`, r) || r == '\r' || r == '\n' {
			return '_'
		}
		return r
	}, strings.Join(parts, "-"))
	if len(number) > buildNumberLimit {
		number = number[:buildNumberLimit]
	}
	return strings.TrimRight(number, ".")
}
