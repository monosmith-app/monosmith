package release

import (
	"fmt"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/pubpackage"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// PubWorkflow is the workflow file the release starts to publish to pub.dev.
const PubWorkflow = "pub.yml"

// DefaultPubTag is the tag pub.dev's automated publishing matches when a
// release sets none: pub.dev's tag patterns cannot hold the "@" of smith's
// release tags.
const DefaultPubTag = "{name}-v{version}"

// PubTag is the tag a project's pub.dev publish runs on, for a version.
func PubTag(p *projectgraph.Project, version string) string {
	pattern := DefaultPubTag
	if p.Release != nil && p.Release.PubTag != "" {
		pattern = p.Release.PubTag
	}
	return strings.NewReplacer("{name}", pubName(p), "{version}", version).Replace(pattern)
}

// pubName is the project's name in its pub tag: its tag prefix without the
// "@" of smith's own tags, or without the "-v" of a prefix that already is
// the pub tag's form (`acme-v`), so such a project's release tag, `acme-v1.4.0`,
// is the pub tag and not `acme-v-v1.4.0`.
func pubName(p *projectgraph.Project) string {
	prefix := TagPrefix(p)
	if name := strings.TrimSuffix(prefix, "-v"); name != prefix && name != "" {
		return name
	}
	return strings.TrimSuffix(prefix, "@")
}

// PubPackage is the pub.dev package a project is published as: an app's
// configured one, or a Dart package's own name when its pubspec publishes
// to pub.dev. Empty when the project is not published there.
func PubPackage(p *projectgraph.Project) string {
	switch {
	case p.Release != nil && p.Release.Pub != "":
		return p.Release.Pub
	case p.Kind == protocol.ProjectPackage && p.PublishTo == "pub.dev" && len(p.Exports) > 0:
		return p.Exports[0]
	}
	return ""
}

// dispatcher starts the pub workflow; tests replace it.
var dispatcher = func(repo githubrelease.Repository, ref string, inputs map[string]string) error {
	client, err := githubrelease.FromEnvironment()
	if err != nil {
		return err
	}
	return client.DispatchWorkflow(repo, PubWorkflow, ref, inputs)
}

// pubHas reports whether pub.dev has a version; tests replace it.
var pubHas = pubpackage.Has

// pubTagElsewhere reports the commit an existing pub tag is on when it is not
// commit, the one the pub run must start from: a workflow run uses the
// workflow file and the smith pinned in the commit it starts from.
func pubTagElsewhere(s *session.Session, tag, commit string) (string, bool) {
	existing := s.Git.TagCommit(tag)
	return existing, existing != "" && existing != commit
}

func short(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}

// dispatchPub has a released app published to pub.dev. pub.dev accepts a
// package only from a GitHub Actions run on a tag matching its pattern, and
// a tag pushed with the workflow's own token starts no workflow, so this
// tags the release commit with the pub tag and starts the generated pub
// workflow on it, which builds the package from the GitHub Release's
// archives. Nothing happens when pub.dev already has the version, so a
// re-run completes a release.
func dispatchPub(s *session.Session, p *projectgraph.Project, item Item, res *ReportResult, say func(string, ...any)) error {
	name := PubPackage(p)
	if name == "" || Native(p) {
		return nil
	}
	tag := PubTag(p, item.Next)
	switch {
	case os.Getenv("TF_BUILD") != "" || os.Getenv("GITLAB_CI") != "":
		say("not publishing %s to pub.dev: pub.dev takes packages only from GitHub Actions", name)
		return nil
	case !s.Config.Release.GitHubReleases:
		say("not publishing %s to pub.dev: its workflow builds the package from the GitHub Release; set [release] github_releases = true", name)
		return nil
	case !res.Pushed:
		say("not publishing %s to pub.dev: the tag was not pushed", name)
		return nil
	}
	has, err := pubHas(name, item.Next)
	if err != nil {
		return fmt.Errorf("pub.dev: %w", err)
	}
	if has {
		say("pub.dev already has %s %s", name, item.Next)
		res.ExistingPackages = append(res.ExistingPackages, name+"@"+item.Next)
		return nil
	}
	commit := s.Git.TagCommit(item.Tag)
	if elsewhere, ok := pubTagElsewhere(s, tag, commit); commit != "" && ok {
		return fmt.Errorf("pub.dev: the tag %s already exists on %s, not on the release commit %s (%s); the pub run would start from that old commit, with its old workflow. Move the tag or set another pub_tag, then re-run the release", tag, short(elsewhere), short(commit), item.Tag)
	}
	if s.Git.TagCommit(tag) == "" {
		if commit == "" {
			return fmt.Errorf("pub.dev: %s has no commit to tag %s on", item.Tag, tag)
		}
		if err := s.Git.CreateTagAt(tag, fmt.Sprintf("%s %s for pub.dev", item.Project, item.Next), commit); err != nil {
			return fmt.Errorf("pub.dev: %w", err)
		}
	}
	if err := s.Git.PushTag(s.Config.Release.Remote, tag); err != nil {
		return fmt.Errorf("pushing %s: %w (re-run the release to publish to pub.dev)", tag, err)
	}
	repo, err := Repository(s)
	if err != nil {
		return err
	}
	if err := dispatcher(repo, "refs/tags/"+tag, map[string]string{"project": item.Project, "version": item.Next}); err != nil {
		return fmt.Errorf("starting %s on %s: %w (allow the workflow actions: write, then re-run the release)", PubWorkflow, tag, err)
	}
	res.PubDispatch = fmt.Sprintf("%s %s from %s", name, item.Next, tag)
	say("dispatched pub.dev publish for %s %s (%s on %s)", name, item.Next, PubWorkflow, tag)
	return nil
}
