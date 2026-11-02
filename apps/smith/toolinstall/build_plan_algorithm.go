package toolinstall

import "slices"

// OptionKind says what an option does.
type OptionKind string

// The kinds of option, in the order they are offered.
const (
	// OptionInstall runs a package manager's install command.
	OptionInstall OptionKind = "install"
	// OptionPage points to the official download page.
	OptionPage OptionKind = "page"
	// OptionMyself leaves the install to the person and prints the ways.
	OptionMyself OptionKind = "myself"
	// OptionSkip goes on without the tool.
	OptionSkip OptionKind = "skip"
)

// Option is one answer to "how do you want to get this tool".
type Option struct {
	Kind OptionKind `json:"kind"`
	// Value is the stable name of the option: "install:winget", "page",
	// "myself", "skip". The --manager flag and the editors use it.
	Value string `json:"value"`
	// Manager is the package manager of an install option.
	Manager string `json:"manager,omitempty"`
	// Command is the install command of an install option.
	Command []string `json:"command,omitempty"`
	// Admin is true when the command needs administrator rights.
	Admin bool `json:"admin,omitempty"`
	// Label and Hint say the option in plain words, with its consequence.
	Label string `json:"label"`
	Hint  string `json:"hint"`
	// Note is a sentence the person needs before choosing, from the plugin.
	Note string `json:"note,omitempty"`
}

// ToolPlan is the presented choice for one tool that is missing or outdated.
type ToolPlan struct {
	Status
	Options []Option `json:"options"`
	// Default is the Value of the marked option.
	Default string `json:"default"`
}

// Find returns the option with a value.
func (p ToolPlan) Find(value string) (Option, bool) {
	for _, o := range p.Options {
		if o.Value == value {
			return o, true
		}
	}
	return Option{}, false
}

// BuildPlan builds the options for one tool. An install option exists for each
// way the plugin declares that a package manager on this machine can run. The
// marked default is the first install option that needs no administrator
// rights, else the first install option; with no install option it is "I will
// install it myself". Nothing here runs.
//
// Parameters: st is a tool that is missing or outdated; managers are the
// managers on the machine; env is the system.
func BuildPlan(st Status, managers []Manager, env Env) ToolPlan {
	plan := ToolPlan{Status: st}
	version := st.Wanted.Version(st.Suggested)
	for _, in := range st.Install {
		if !slices.ContainsFunc(managers, func(m Manager) bool { return m.Name == in.Manager }) {
			continue
		}
		cmd, ok := BuildCommand(in.Manager, Fill(in.Package, version), env)
		if !ok {
			continue
		}
		hint := "runs " + cmd.String()
		if cmd.Admin {
			hint += "; " + adminSentence(in.Manager)
		}
		plan.Options = append(plan.Options, Option{
			Kind: OptionInstall, Value: "install:" + in.Manager, Manager: in.Manager,
			Command: cmd.Argv, Admin: cmd.Admin, Label: "Install with " + in.Manager, Hint: hint, Note: in.Note,
		})
	}
	plan.Default = "myself"
	for _, o := range plan.Options {
		if !o.Admin {
			plan.Default = o.Value
			break
		}
	}
	if plan.Default == "myself" && len(plan.Options) > 0 {
		plan.Default = plan.Options[0].Value
	}
	plan.Options = append(plan.Options,
		Option{Kind: OptionPage, Value: "page", Label: "Use the official page", Hint: "go to " + st.Download + " and install it there"},
		Option{Kind: OptionMyself, Value: "myself", Label: "I will install it myself", Hint: "smith shows the ways and goes on; run smith tools to check later"},
		Option{Kind: OptionSkip, Value: "skip", Label: "Skip it", Hint: "go on without it; " + st.Consequence},
	)
	return plan
}

// adminSentence says in plain words why an install needs approval.
func adminSentence(manager string) string {
	if manager == "winget" {
		return "Windows may ask you to approve it as administrator"
	}
	return "it needs administrator rights, and sudo asks for your password"
}
