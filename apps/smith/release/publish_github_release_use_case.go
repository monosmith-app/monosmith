package release

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/sbom"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// publishGitHubRelease creates the tag's GitHub Release and attaches the
// assets. When assets is nil (a resumed release), the plugin lists the
// expected files and they are rebuilt only if the release lacks any.
func publishGitHubRelease(ctx context.Context, s *session.Session, p *projectgraph.Project, item Item, res *ReportResult, say func(string, ...any), assets []string) error {
	if !res.Pushed {
		say("not creating a GitHub Release: the tag was not pushed")
		if PacksToRelease(p) {
			say("the package files are in %s; push the tag and run the release again to attach them", PackDir(p.Name))
		}
		return nil
	}
	client, err := githubrelease.FromEnvironment()
	if err != nil {
		return err
	}
	repo, err := Repository(s)
	if err != nil {
		return err
	}
	if assets == nil && (p.Release != nil || attachesItsFile(p)) && !Native(p) && !BuildsAssetsFromCommand(p) {
		existing, found, err := client.ReleaseByTag(repo, item.Tag)
		if err != nil {
			return err
		}
		planned, err := publish(ctx, s, p, item, true)
		if err != nil {
			return err
		}
		packages, err := linuxPackages(s, p, item, planned, true)
		if err != nil {
			return err
		}
		if !found || missing(existing, append(planned.Assets, packages...)) {
			say("building assets for %s", item.Next)
			built, err := publish(ctx, s, p, item, false)
			if err != nil {
				return err
			}
			packages, err := linuxPackages(s, p, item, built, false)
			if err != nil {
				return err
			}
			assets = append(built.Assets, packages...)
		}
	}
	if len(assets) > 0 && p.Release != nil && !Native(p) {
		bom, err := sbom.Write(s, p, sbom.Request{Version: item.Next,
			Product: p.Release.Product, OutDir: "dist/release/" + p.Name, SmithVersion: s.Version})
		if err != nil {
			return fmt.Errorf("SBOM for %s: %w (re-run the release to finish it)", item.Tag, err)
		}
		if bom != "" {
			assets = append(slices.Clone(assets), bom)
		}
	}
	files := make([]string, len(assets))
	for i, a := range assets {
		files[i] = filepath.Join(s.Root, filepath.FromSlash(a))
	}
	out, err := githubrelease.Ensure(client, repo, githubrelease.Request{Tag: item.Tag, Title: item.Project + " " + item.Next, Notes: RenderNotes(item), Prerelease: strings.Contains(item.Next, "-"), Assets: files, Verify: PacksToRelease(p)})
	if err != nil {
		return fmt.Errorf("GitHub Release %s: %w (re-run the release to finish it)", item.Tag, err)
	}
	res.GitHubRelease = out.URL
	if out.Created {
		say("created GitHub Release %s", out.URL)
	}
	if len(out.Replaced) > 0 {
		say("replaced %s: the attached file differed from the one packed now", strings.Join(out.Replaced, ", "))
	}
	if len(out.Uploaded) > 0 {
		say("attached %s", strings.Join(out.Uploaded, ", "))
	}
	return nil
}
