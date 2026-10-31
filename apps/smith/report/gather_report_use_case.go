package report

import (
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/graphview"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// Gather reads the open workspace into Data. A part that fails (the release
// plan on a repository with no history, the derive plan on a broken
// smith.lock) is reported in its section instead of failing the whole page.
func Gather(s *session.Session) Data {
	d := Data{Workspace: s.Config.Workspace.Name, Branch: s.Config.Release.Branch, GraphSVG: graphview.SVG(s.Graph, nil)}
	if d.Branch == "" {
		d.Branch = "main"
	}
	for _, p := range s.Graph.Projects {
		pr := Project{Name: p.Name, Root: p.Root, Kind: string(p.Kind), Language: p.Language, Stack: p.Stack, Conventions: p.Conventions, Deps: p.Deps}
		if st, ok := s.Config.Stacks[p.Stack]; ok {
			pr.Choices = st.Choices
		}
		d.Projects = append(d.Projects, pr)
	}
	sort.Slice(d.Projects, func(i, j int) bool { return d.Projects[i].Name < d.Projects[j].Name })
	for _, c := range s.Graph.Cycles() {
		d.Cycles = append(d.Cycles, strings.Join(c, " → "))
	}
	d.Health = healthOf(doctor.Run(s))
	d.Drift = driftOf(s)
	d.Release = releaseOf(s)
	return d
}

// healthOf maps the doctor report: failures first, then warnings, then passes.
func healthOf(r doctor.Report) Health {
	h := Health{Passed: r.Passed, Warned: r.Warnings, Failed: r.Failed}
	for _, f := range r.Findings {
		c := Check{Name: f.Check, Detail: f.Detail, Remedy: f.Remedy, Status: StatusFail}
		switch {
		case f.OK:
			c.Status = StatusPass
		case f.Warning:
			c.Status = StatusWarn
		}
		h.Checks = append(h.Checks, c)
	}
	rank := map[string]int{StatusFail: 0, StatusWarn: 1, StatusPass: 2}
	sort.SliceStable(h.Checks, func(i, j int) bool { return rank[h.Checks[i].Status] < rank[h.Checks[j].Status] })
	return h
}

// driftOf builds the derive plan when derived files are on.
func driftOf(s *session.Session) Drift {
	d := Drift{Enabled: s.Config.Derive.Enabled}
	if !d.Enabled {
		return d
	}
	plan, _, err := derive.BuildPlan(s)
	if err != nil {
		d.Error = err.Error()
		return d
	}
	for _, i := range plan.Items {
		if !i.NeedsAction() {
			d.Current++
			continue
		}
		d.Items = append(d.Items, DriftItem{Path: i.File.Path, Renderer: i.File.Renderer, Status: string(i.Status), Owner: i.Owner, Detail: i.Detail})
	}
	return d
}

// releaseOf plans the release of the current commit, as if it were merged to
// the release branch. It does not ask a registry, so it can say less than
// `smith release --dry-run`: the check of the public API is left out.
func releaseOf(s *session.Session) Release {
	plan, err := release.MakePlan(s.Git, s.Graph)
	if err != nil {
		return Release{Error: err.Error()}
	}
	r := Release{Unchanged: plan.Unchanged}
	for _, i := range plan.Items {
		r.Items = append(r.Items, ReleaseItem{Project: i.Project, Current: i.Current, Next: i.Next, Level: i.Level, Reason: i.Reason})
	}
	return r
}
