package release

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// recordDeployments shows each registry a release published to in the
// repository's Deployments sidebar, on GitHub Actions only. A local run, another
// CI, a dry run and [release] github_deployments = false record nothing. A
// failure only prints a warning: the packages are published and the release is
// done.
func recordDeployments(s *session.Session, report Report, out io.Writer) {
	if report.DryRun || os.Getenv("GITHUB_ACTIONS") != "true" || !s.Config.Release.DeploymentsOn() || len(deployable(report)) == 0 {
		return
	}
	client, err := githubrelease.FromEnvironment()
	if err != nil {
		deploymentWarning(out, err)
		return
	}
	repo, err := Repository(s)
	if err != nil {
		deploymentWarning(out, err)
		return
	}
	recordWith(client, repo, report, out)
}

// deployable lists the deployments a report holds: a registry package the
// release published now, from a tag that is on GitHub.
func deployable(report Report) []githubrelease.Deployment {
	var all []githubrelease.Deployment
	for _, res := range report.Results {
		if res.Status != "released" || !res.Pushed {
			continue
		}
		for _, t := range TargetsOf(res) {
			page := packagePage(t.Kind, t.pkg, res.Next)
			if page == "" || t.fresh == 0 || t.pkg == "" {
				continue
			}
			all = append(all, githubrelease.Deployment{Environment: t.Kind, Ref: res.Tag, Description: t.pkg + " " + res.Next, URL: page})
		}
	}
	return all
}

// recordWith records the report's deployments with a client; the first failure
// stops the rest, because the same missing permission would fail each of them.
func recordWith(client *githubrelease.Client, repo githubrelease.Repository, report Report, out io.Writer) {
	for _, d := range deployable(report) {
		if err := client.RecordDeployment(repo, d); err != nil {
			deploymentWarning(out, fmt.Errorf("%s %s: %w", d.Environment, d.Description, err))
			return
		}
	}
}

// deploymentWarning is the one line a failed call prints, with its remedy.
func deploymentWarning(out io.Writer, err error) {
	reason, _, _ := strings.Cut(err.Error(), "\n")
	if len(reason) > 160 {
		reason = reason[:160] + "…"
	}
	_, _ = fmt.Fprintf(out, "warning: could not record the release as GitHub deployments: %s (give the job permissions: deployments: write, or set github_deployments = false under [release] in smith.toml)\n", reason)
}
