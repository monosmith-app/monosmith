package composeview

import (
	"context"
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/toolinstall"
	"github.com/monosmith-app/monosmith/apps/smith/tui/languagenames"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Options are what `smith compose` runs with.
type Options struct {
	// Search is the words to find; empty lists everything.
	Search string
	// Lang is the language asked for (--lang).
	Lang string
	// With are the choices to check; empty means only the catalogue.
	With []string
	// Kind and Name describe the project the plan is for.
	Kind, Name string
}

// Report is the JSON `smith compose --json` prints.
type Report struct {
	Layers []compose.Layer `json:"layers"`
	// Items are the catalogue, or the items that matched Search, best first.
	Items  []compose.Hit `json:"items"`
	Search string        `json:"search,omitempty"`
	// Outcomes are the answers to "what are you building?" (a website, a
	// function app...), from the whole catalogue: the editors list them as the
	// first question of Add, as the terminal screen does.
	Outcomes []compose.Outcome `json:"outcomes"`
	// Check and Plan are set when --with is given.
	Check *compose.Check `json:"check,omitempty"`
	Plan  *compose.Plan  `json:"plan,omitempty"`
}

// Source reads the catalogue's source from the loaded plugins: the languages
// that can generate, the choices, and the structure and deploy plugins.
//
// Parameters: s is the open session.
// Returns: the source.
func Source(s *session.Session) compose.Source {
	src := compose.Source{Offers: s.Plugins.Offers()}
	for _, c := range s.Plugins.Clients {
		name := c.Info.Plugin.Name
		if c.Info.Capabilities.Generate && pluginhost.HasRole(c, protocol.RoleLanguage) {
			src.Languages = append(src.Languages, compose.Language{Name: name, Title: languagenames.Of(name), Description: languagenames.Describe(name)})
		}
		for role, layer := range map[protocol.Role]string{protocol.RoleStructure: "architecture", protocol.RoleDeploy: "ci"} {
			if pluginhost.HasRole(c, role) {
				src.Plugins = append(src.Plugins, compose.WorkspacePlugin{Name: name, Layer: layer, Description: string(role) + " plugin" + languagesOf(c.Info.Languages)})
			}
		}
	}
	return src
}

// languagesOf words the languages a plugin serves.
func languagesOf(languages []string) string {
	if len(languages) == 0 {
		return ", for every language"
	}
	return ", for " + strings.Join(languages, ", ")
}

// toolsOf maps the detected tools of a selection to the plan's tools: the
// state, and the line that offers the install. A tool that a choice limits
// names that choice; the others name the language.
func toolsOf(statuses []toolinstall.Status, language string) []compose.Tool {
	out := make([]compose.Tool, 0, len(statuses))
	for _, st := range statuses {
		needed := st.Choice
		if needed == "" {
			needed = language
		}
		state := "installed"
		switch st.State {
		case toolinstall.Missing:
			state = "missing"
		case toolinstall.Outdated:
			state = "outdated"
		}
		out = append(out, compose.Tool{Choice: needed, Name: st.Name, Title: st.Title, Binary: st.ToolNeed.Program, State: state, Command: "smith tools install " + st.Name})
	}
	return out
}

// Run makes the report: the catalogue (searched), and, for --with, the check
// and the plan. It writes nothing.
//
// Parameters: s is the open session; o are the options.
// Returns: the report. It returns an error for a --lang no plugin handles.
func Run(s *session.Session, o Options) (Report, error) {
	cat := compose.Build(Source(s))
	rep := Report{Layers: cat.Layers, Items: compose.Search(cat.Items, o.Search), Search: o.Search, Outcomes: compose.Outcomes(cat.Items)}
	language := ""
	if o.Lang != "" {
		client, err := s.Plugins.ForLanguage(o.Lang)
		if err != nil {
			return Report{}, err
		}
		language = client.Info.Plugin.Name
		rep.Items = onlyLanguage(rep.Items, language)
	}
	if len(o.With) == 0 {
		return rep, nil
	}
	offers := s.Plugins.Offers()
	check := compose.CheckPicks(offers, language, o.With)
	rep.Check = &check
	in := compose.PlanInput{Kind: o.Kind, Name: o.Name, Check: check}
	if check.Complete {
		in.Tools = toolsOf(toolinstall.SelectionStatuses(context.Background(), toolinstall.ExecDetector{}, s.Root, toolinstall.SessionSources(s), check.Language, check.Selection), check.Language)
		sel, err := choiceselect.SolveIn(offers, check.Language, o.With)
		if err != nil {
			return Report{}, err
		}
		st := sel.Stack("", "")
		in.Language, in.Scaffold, in.Toolchain = st.Language, st.Scaffold, st.Toolchain
		in.Files = append(choiceLines(cat, check), addPlanLines(scaffold.PlanAdd(s, scaffold.AddOptions{Kind: kindOr(o.Kind), Name: nameOr(o.Name), Lang: check.Language, With: check.Selection}))...)
	}
	plan := compose.BuildPlan(in)
	rep.Plan = &plan
	return rep, nil
}

// choiceLines say what each choice of the selection puts in the project, in
// the words its plugin gave, because the exact files belong to the plugins.
func choiceLines(cat compose.Catalogue, check compose.Check) []compose.PlanLine {
	var out []compose.PlanLine
	for _, name := range check.Selection {
		if it, ok := cat.Find(check.Language + "/" + name); ok {
			out = append(out, compose.PlanLine{Action: "choice", Path: it.Title, Note: it.Description})
		}
	}
	return out
}

// onlyLanguage keeps the items of a language and those of no language.
func onlyLanguage(hits []compose.Hit, language string) []compose.Hit {
	var out []compose.Hit
	for _, h := range hits {
		if h.Language == "" || h.Language == language || h.Kind == compose.KindLanguage && h.Name == language {
			out = append(out, h)
		}
	}
	return out
}

func kindOr(kind string) string {
	if kind == "" {
		return "app"
	}
	return kind
}

func nameOr(name string) string {
	if name == "" {
		return "my-app"
	}
	return name
}

// addPlanLines flattens the plan of smith add into the lines the composer shows.
func addPlanLines(p operationplan.Plan) []compose.PlanLine {
	var out []compose.PlanLine
	for _, g := range p.Groups {
		for _, l := range g.Lines {
			note := l.Note
			if note == "" {
				note = g.Heading
			}
			out = append(out, compose.PlanLine{Action: string(l.Action), Path: l.Path, Note: note})
		}
	}
	return out
}

// Refusal is the error for a run that needs a terminal and has none.
func Refusal() error {
	return fmt.Errorf("smith compose is a screen and needs a terminal. To read the catalogue without one, add --json (smith compose --json --search desktop). To check a selection, add --with a,b,c (smith compose --json --lang node --with react)")
}
