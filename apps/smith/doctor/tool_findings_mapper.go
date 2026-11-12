package doctor

import (
	"context"
	"github.com/monosmith-app/monosmith/apps/smith/doctorchecks"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/toolinstall"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// checkName reduces a check to the words that identify it, so the tool check
// "a JDK is installed" and a plugin's "JDK is installed" match: case and a
// leading "the ", "a " or "an " are ignored.
func checkName(check string) string {
	c := strings.ToLower(strings.TrimSpace(check))
	for _, article := range []string{"the ", "a ", "an "} {
		c = strings.TrimPrefix(c, article)
	}
	return c
}

// ToolFindings adds the missing tools the workspace needs to the findings. A
// plugin that already reports the same tool keeps its finding and gains the
// fix; otherwise a warning is added. The fix is a choice with the same options
// as smith tools install: nothing runs until the person picks (ADR 0077).
//
// Parameters: s is the open session; findings are the findings so far.
// Returns: the findings with the tool checks.
func ToolFindings(ctx context.Context, s *session.Session, findings []Finding) []Finding {
	list := toolinstall.DetectAll(ctx, toolinstall.ExecDetector{}, toolinstall.WorkspaceNeeds(s, nil, nil))
	plans := map[string]toolinstall.ToolPlan{}
	for _, p := range toolinstall.PlansFor(list, false) {
		plans[p.Name] = p
	}
	for _, st := range list {
		check := st.Title + " is installed"
		plan, lacking := plans[st.Name]
		if !lacking || st.Exact {
			// A tool with one wanted version (golangci-lint) has its own check in the
			// plugin, which says where it found the program and why a fix may not help.
			continue
		}
		fix := ToolFix(plan)
		attached := false
		for i := range findings {
			if checkName(findings[i].Check) == checkName(check) && !findings[i].OK {
				if findings[i].Fix == nil {
					findings[i].Fix = &fix
				}
				if st.OffPath != "" {
					// The plugin only knows the program is not on PATH; smith found it.
					findings[i].Problem = st.Title + " " + st.Problem()
					findings[i].Remedy = "smith tools use " + st.Name
				}
				attached = true
			}
		}
		if attached {
			continue
		}
		detail := st.Why
		problem := st.Title + " is not on PATH"
		remedy := "smith tools install " + st.Name
		if st.OffPath != "" {
			problem, remedy = st.Title+" "+st.Problem(), "smith tools use "+st.Name
		}
		if st.State == toolinstall.Outdated {
			detail = st.Title + " " + st.Version + " is installed, " + st.Wanted.String() + " is wanted"
			problem = detail
		}
		findings = append(findings, Finding{Source: "smith", Finding: protocol.Finding{
			Check: check, OK: false, Warning: true, Area: doctorchecks.AreaToolchain,
			Problem: problem,
			Detail:  detail + "; if you skip it, " + st.Consequence,
			Remedy:  remedy,
			Fix:     &fix,
		}})
	}
	return withoutRepeatedMachineChecks(findings)
}

// withoutRepeatedMachineChecks keeps one finding per tool. Each project of a
// kind asks its plugin about the same program (two function apps both ask for
// Azure Functions Core Tools), and the machine has one answer, so the list
// shows it once.
func withoutRepeatedMachineChecks(findings []Finding) []Finding {
	seen := map[string]int{}
	out := findings[:0:0]
	for _, f := range findings {
		if IsMachineCheck(f) {
			key := checkName(f.Check) + "|" + strconv.FormatBool(f.OK)
			if at, ok := seen[key]; ok {
				// The kept finding now speaks for several projects, so it is
				// about the machine and no longer about the first project.
				if kept := &out[at]; kept.Project != "" && kept.Project != f.Project {
					kept.Project, kept.Area = "", doctorchecks.AreaToolchain
				}
				continue
			}
			seen[key] = len(out)
		}
		out = append(out, f)
	}
	return out
}

// ToolFix turns the plan of a tool into a fix with one choice. The marked
// default is the first option. A command that needs administrator rights is
// two options: show it so the person runs it (first), or let smith run it (the
// system asks for its own permission).
func ToolFix(plan toolinstall.ToolPlan) protocol.Fix {
	var first, rest []protocol.FixChoiceOption
	for _, o := range plan.Options {
		var group []protocol.FixChoiceOption
		cmd := strings.Join(o.Command, " ")
		switch {
		case o.ChangesPath():
			// The folder answers: the label says what changes, not the command.
			group = append(group, toolOption(o.Label+": "+o.Hint, o.Label+": "+plan.Title, o.Command))
		case o.Kind == toolinstall.OptionMyself && plan.OffPath != "":
			group = append(group, toolOption(o.Label, toolinstall.RenderFolder(plan), nil))
		case o.Kind == toolinstall.OptionInstall && o.Admin:
			group = append(group,
				toolOption("I will run it myself: "+cmd+" (needs administrator rights)", "Run this yourself, then open a new terminal: "+cmd, nil),
				toolOption("Let smith run it: "+cmd+" (the system asks for your permission)", "Install "+plan.Title+" with "+o.Manager+": "+cmd, o.Command))
		case o.Kind == toolinstall.OptionInstall:
			group = append(group, toolOption(o.Label+": "+cmd, "Install "+plan.Title+" with "+o.Manager+": "+cmd, o.Command))
		default:
			group = append(group, toolOption(o.Label+": "+o.Hint, o.Label+": "+plan.Title, nil))
		}
		if o.Value == plan.Default {
			first = append(first, group...)
		} else {
			rest = append(rest, group...)
		}
	}
	return protocol.Fix{
		Description: "Get " + plan.Title,
		Choices:     []protocol.FixChoice{{Question: "How do you want to get " + plan.Title + "? " + plan.Why + ".", Options: append(first, rest...)}},
	}
}

// toolOption is one answer of the fix: argv is the command it runs, or nil
// when the answer runs nothing.
func toolOption(label, description string, argv []string) protocol.FixChoiceOption {
	fix := protocol.Fix{Description: description}
	if argv != nil {
		fix.Steps = []protocol.FixStep{{Description: description, Argv: argv}}
	}
	return protocol.FixChoiceOption{Label: label, Fix: fix}
}
