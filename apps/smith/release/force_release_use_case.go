package release

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// ForceInfo is what a forced release needs to know about one package: the
// facts that plan building reads from git and the project graph.
type ForceInfo struct {
	// Plugin is the plugin that releases the package.
	Plugin string
	// TagPrefix starts the package's release tags.
	TagPrefix string
	// Latest is the last released version; on a channel, the last of the channel.
	Latest Version
	// Released is false when the package has no release yet.
	Released bool
	// OnChannel is true when Latest is a release of the channel.
	OnChannel bool
	// LatestTag is the tag of Latest.
	LatestTag string
}

// ForceUsage is what smith prints when --force has no bump to go with it.
const ForceUsage = "--force needs the package and the level: smith release --bump <project>=patch|minor|major --force"

// ForcePlan adds to a plan the packages the person chose to release although no
// commit releases them: a manual (forced) release of the same code under a new
// version. A name that is not in plan.Unchanged is left to Adjust, so --force
// on a package that already has a releasing commit only applies the bump.
//
// Parameters: plan is what the commits asked for; bumps is --bump (project to
// level); order is the dependency order of every project; info reads one
// package's facts; ch is the release channel of the branch (nil: stable).
// Returns: the plan with the forced items in dependency order, the bumps that
// Adjust still has to apply, and one note per forced package. An error for a
// level smith does not know or a package that was never released.
func ForcePlan(plan Plan, bumps map[string]string, order []string, info func(name string) (ForceInfo, bool), ch *Channel) (Plan, map[string]string, []string, error) {
	out := plan
	out.Items = slices.Clone(plan.Items)
	rest := map[string]string{}
	var notes []string
	for _, name := range sortedKeys(bumps) {
		fi, known := info(name)
		if !slices.Contains(plan.Unchanged, name) || !known {
			rest[name] = bumps[name]
			continue
		}
		level, ok := levelNamed(bumps[name])
		if !ok || level == None {
			return plan, nil, nil, fmt.Errorf("--bump %s=%s: the level is patch, minor or major", name, bumps[name])
		}
		if !fi.Released {
			return plan, nil, nil, fmt.Errorf("--bump %s --force: %s was never released, so there is no version to bump; set it with a version(%s)[x.y.z] commit", name, name, name)
		}
		item, differs := forcedItem(name, fi, level, ch)
		out.Items = append(out.Items, item)
		out.Unchanged = slices.DeleteFunc(slices.Clone(out.Unchanged), func(n string) bool { return n == name })
		notes = append(notes, fmt.Sprintf("%s: no commit since %s releases it; --force releases the same code as %s -> %s (%s bump)", name, fi.LatestTag, item.Current, item.Next, level))
		if differs != "" {
			notes = append(notes, fmt.Sprintf("%s: you asked for %s, and the version is %s -> %s: %s", name, level, item.Current, item.Next, differs))
		}
	}
	position := func(i Item) int { return slices.Index(order, i.Project) }
	slices.SortStableFunc(out.Items, func(a, b Item) int { return position(a) - position(b) })
	return out, rest, notes, nil
}

// forcedItem builds the plan item of a forced release. The second result says
// when the level moves the version differently from what the person asked.
func forcedItem(name string, fi ForceInfo, level Level, ch *Channel) (Item, string) {
	var next Version
	differs := ""
	if ch != nil {
		next = ch.Next(fi.Latest, fi.OnChannel)
	} else {
		next, differs = fi.Latest.BumpAsked(level)
	}
	return Item{
		Project: name,
		Plugin:  fi.Plugin,
		Current: fi.Latest.String(),
		Next:    next.String(),
		Level:   level.String(),
		Tag:     fi.TagPrefix + next.String(),
		Manual:  true,
		Reason:  ManualReason(fi.LatestTag),
	}, differs
}

// ManualReason is the reason shown for a package released without a releasing
// commit.
//
// Parameters: lastTag is the package's last release tag.
// Returns: one sentence.
func ManualReason(lastTag string) string {
	return fmt.Sprintf("manual release, forced: no commit since %s releases this: you are releasing the same code under a new version", lastTag)
}

// ManualNames lists the packages of a plan that release without a releasing
// commit, in plan order.
func ManualNames(plan Plan) []string {
	var names []string
	for _, i := range plan.Items {
		if i.Manual {
			names = append(names, i.Project)
		}
	}
	return names
}

// ManualConsequences says in plain words what a manual release does beyond a
// normal one. Empty when the plan has no manual release.
func ManualConsequences(plan Plan) string {
	names := ManualNames(plan)
	if len(names) == 0 {
		return ""
	}
	return fmt.Sprintf("%s has no commit that releases it. The new tag points at HEAD, and the same code is published under a new version, with a GitHub Release and the assets.",
		strings.Join(names, ", "))
}

// ForceFromSession reads the facts of ForcePlan from git and the project graph.
//
// Parameters: git and graph are the session's; ch is the branch's channel.
// Returns: the lookup ForcePlan needs and the dependency order.
func ForceFromSession(git *gitclient.Git, graph *projectgraph.Graph, ch *Channel) (func(string) (ForceInfo, bool), []string, error) {
	order, err := graph.Order()
	if err != nil {
		return nil, nil, err
	}
	git = git.WithRefCache()
	info := func(name string) (ForceInfo, bool) {
		p := graph.Get(name)
		if p == nil || !p.Releases() {
			return ForceInfo{}, false
		}
		latest, tag, released := LatestTag(git, p)
		fi := ForceInfo{Plugin: p.Plugin, TagPrefix: TagPrefix(p), Latest: latest, Released: released, LatestTag: tag}
		if ch != nil {
			if v, t, ok := latestChannelTag(git, p, *ch); ok {
				fi.Latest, fi.LatestTag, fi.Released, fi.OnChannel = v, t, true, true
			}
		}
		return fi, true
	}
	return info, order, nil
}

// ErrForceWithoutBump is returned when --force is given with no --bump.
var ErrForceWithoutBump = errors.New(ForceUsage)

// TagMessage is the message of a release tag. A manual (forced) release says so.
//
// Parameters: item is the release.
// Returns: the annotated tag's message.
func TagMessage(item Item) string {
	msg := fmt.Sprintf("%s %s", item.Project, item.Next)
	if item.Manual {
		msg += " (manual release: no commit released it, the same code under a new version)"
	}
	return msg
}

// AdjustPlan applies --hold, --bump and --force to the plan of the session's
// branch: a forced package joins the plan first, then Adjust applies the rest.
//
// Parameters: s is the session; plan is what the commits asked for; hold and
// bumpFlags are the flag values; force is --force.
// Returns: the adjusted plan and one note per change. An error for a bad flag
// value, --force without --bump, or a name Adjust or ForcePlan refuses.
func AdjustPlan(s *session.Session, plan Plan, hold, bumpFlags []string, force bool) (Plan, []string, error) {
	bumps, err := ParseBumps(bumpFlags)
	if err != nil {
		return plan, nil, err
	}
	if force && len(bumps) == 0 {
		return plan, nil, ErrForceWithoutBump
	}
	var notes []string
	if force {
		ch := ChannelOf(s.Config.Release, PlanBranch(s))
		info, order, err := ForceFromSession(s.Git, s.Graph, ch)
		if err != nil {
			return plan, nil, err
		}
		if plan, bumps, notes, err = ForcePlan(plan, bumps, order, info, ch); err != nil {
			return plan, nil, err
		}
	}
	plan, more, err := Adjust(plan, Adjustments{Hold: hold, Bump: bumps})
	return plan, append(notes, more...), err
}
