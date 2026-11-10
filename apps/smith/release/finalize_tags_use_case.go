package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/npmbinaries"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/npmpropagation"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// pendingMove is a dist-tag a project's publish left under "pending": who moves
// it (a plugin, or the host itself when Plugin is empty) and what to move.
type pendingMove struct {
	Plugin string
	Move   protocol.TagMove
}

// addPending records dist-tags that wait for the last phase of the release and
// keeps the exact commands that move them by hand, in case the release stops
// before that phase.
//
// Parameters: plugin is the plugin that moves them, "" for the host; moves are the tags.
func (r *ReportResult) addPending(plugin string, moves ...protocol.TagMove) {
	for _, m := range moves {
		r.moves = append(r.moves, pendingMove{Plugin: plugin, Move: m})
		r.PendingTags = append(r.PendingTags, m.Command())
	}
}

// tagMover moves the dist-tags of one project: it waits until the packages
// they need are fetchable, then moves each tag.
type tagMover interface {
	// Move returns the tags it moved, the moves still waiting when the time ran
	// out, and one note per waiting move naming the command that does it by hand.
	Move(project string, plugin string, moves []protocol.TagMove) (moved []string, pending []protocol.TagMove, notes []string, err error)
}

// FinalizeTags is the last phase of a release. After every project has
// published, it waits for each pending dist-tag's packages and moves the tag;
// a tag whose packages do not appear in time stays under "pending" with its
// command in the report. A wait here blocks no publish. It runs on a re-run
// too, so a release that stopped on the wait is finished by running it again.
func FinalizeTags(ctx context.Context, s *session.Session, report *Report, out func(format string, args ...any)) error {
	return finalizeTags(report, &sessionMover{ctx: ctx, s: s, out: out}, out)
}

// finalizeTags moves the pending dist-tags of every result with mover.
func finalizeTags(report *Report, mover tagMover, say func(format string, args ...any)) error {
	announced := false
	for i := range report.Results {
		res := &report.Results[i]
		if len(res.moves) == 0 {
			continue
		}
		if !announced {
			announced = true
			say("moving the dist-tags: waiting until npm serves the platform packages")
		}
		var rest []pendingMove
		for _, group := range groupByPlugin(res.moves) {
			moved, pending, notes, err := mover.Move(res.Project, group.plugin, group.moves)
			for _, m := range moved {
				say("%s: moved the dist-tag: %s", res.Project, m)
			}
			for _, n := range notes {
				say("%s: note: %s", res.Project, n)
			}
			for _, m := range pending {
				rest = append(rest, pendingMove{Plugin: group.plugin, Move: m})
			}
			if err != nil {
				res.settle(append(rest, notDone(res.moves, group)...))
				report.Error = res.Project + ": " + err.Error()
				return fmt.Errorf("%s: %w", res.Project, err)
			}
		}
		res.settle(rest)
	}
	return nil
}

// settle keeps only the moves still waiting, and their commands.
func (r *ReportResult) settle(rest []pendingMove) {
	r.moves, r.PendingTags = rest, nil
	for _, m := range rest {
		r.PendingTags = append(r.PendingTags, m.Move.Command())
	}
}

type moveGroup struct {
	plugin string
	moves  []protocol.TagMove
}

// groupByPlugin splits moves by who moves them, keeping their order.
func groupByPlugin(moves []pendingMove) []moveGroup {
	var groups []moveGroup
next:
	for _, m := range moves {
		for i := range groups {
			if groups[i].plugin == m.Plugin {
				groups[i].moves = append(groups[i].moves, m.Move)
				continue next
			}
		}
		groups = append(groups, moveGroup{m.Plugin, []protocol.TagMove{m.Move}})
	}
	return groups
}

// notDone lists the moves of the failed group and of the groups after it: a
// failure leaves them to the next run, with their commands.
func notDone(all []pendingMove, failed moveGroup) []pendingMove {
	var rest []pendingMove
	seen := false
	for _, g := range groupByPlugin(all) {
		seen = seen || g.plugin == failed.plugin
		if seen {
			for _, m := range g.moves {
				rest = append(rest, pendingMove{Plugin: g.plugin, Move: m})
			}
		}
	}
	return rest
}

// sessionMover moves the tags for real: through the project's plugin, or with
// npm for the packages of an app that the host publishes itself.
type sessionMover struct {
	ctx context.Context
	s   *session.Session
	out func(format string, args ...any)
}

func (m *sessionMover) Move(project, plugin string, moves []protocol.TagMove) ([]string, []protocol.TagMove, []string, error) {
	p := m.s.Graph.Get(project)
	if p == nil {
		return nil, moves, nil, fmt.Errorf("the project %s is gone", project)
	}
	waitText := PropagationWait(m.s.Config, p)
	wait, err := npmpropagation.ParseWait(waitText)
	if err != nil {
		return nil, moves, nil, err
	}
	if plugin == "" {
		r := npmbinaries.Release{OutDir: "dist/release/" + project + "/npm", PropagationWait: wait, Progress: func(line string) { m.out("%s: %s", project, line) }}
		done, err := npmbinaries.Finalize(m.s.Root, r, moves)
		return done.Moved, done.Pending, done.Notes, err
	}
	var pr protocol.PublishResult
	params := protocol.PublishParams{Project: p.Project, Finalize: moves, PropagationWait: waitText}
	ok, err := m.s.Plugins.CallContext(m.ctx, plugin, protocol.MethodPublish, params, &pr, wait+5*time.Minute)
	if err != nil {
		return nil, moves, nil, fmt.Errorf("moving the dist-tags with %s: %w", plugin, err)
	}
	if !ok {
		return nil, moves, nil, fmt.Errorf("the %s plugin cannot move a dist-tag", plugin)
	}
	return pr.Moved, pr.PendingMoves, pr.Notes, nil
}

// resumeNPM finishes the npm packages of a version whose tag already exists, so
// a re-run of a release that stopped after publishing its packages completes
// them: it publishes the packages the registry lacks (rebuilding their files
// from the sources), and hands the wrappers still under "pending" to the last
// phase, which waits and moves the dist-tag.
func resumeNPM(ctx context.Context, s *session.Session, p *projectgraph.Project, item Item, versions map[string]string, opts Options, res *ReportResult, say func(string, ...any)) error {
	r := p.Release
	if r == nil || len(r.NPM) == 0 || Native(p) {
		return nil
	}
	probe := npmbinaries.Release{Names: r.NPM, Version: item.Next}
	if ch := ChannelOf(s.Config.Release, s.Git.CurrentBranch()); ch != nil {
		probe.Tag = ch.Tag
	}
	var pairs [][2]string
	for _, platform := range r.Platforms {
		goos, goarch, _ := strings.Cut(platform, "/")
		pairs = append(pairs, [2]string{goos, goarch})
	}
	missing, moves := npmbinaries.Incomplete(s.Root, probe, pairs)
	if len(missing) == 0 {
		res.addPending("", moves...)
		return nil
	}
	say("npm does not have %s yet: publishing what is missing", strings.Join(missing, ", "))
	_, restore, err := stampVersion(ctx, s, p, item, versions, opts, res, say)
	if err != nil {
		return err
	}
	defer restore()
	pr, err := publish(ctx, s, p, item, false)
	if err != nil {
		return err
	}
	return distribute(s, p, item, pr, false, res, say)
}
