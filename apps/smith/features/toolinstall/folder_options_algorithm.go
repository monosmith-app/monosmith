package toolinstall

import "github.com/monosmith-app/monosmith/libs/words/words"

// The managers of the two options that put a found folder on PATH. They are
// install options, so --manager names them as it names a package manager.
const (
	// SmithPathManager lets smith use the folder for its own work (smith tools use).
	SmithPathManager = "smith-path"
	// UserPathManager adds the folder to the Path of the Windows user.
	UserPathManager = "user-path"
)

// ChangesPath reports whether the option puts a found folder on a PATH instead
// of installing a package.
func (o Option) ChangesPath() bool {
	return o.Kind == OptionInstall && (o.Manager == SmithPathManager || o.Manager == UserPathManager)
}

// folderPlan is the choice for a tool that is installed in a folder that is not
// on PATH (st.OffPath): let smith use it from there (the default), add the
// folder to the user PATH (Windows), show the folder and change nothing, or
// skip. A new install would not help, so no package manager is offered.
//
// Parameters: st has OffPath set; env is the system and names the file where
// smith records the folder.
// Returns: the plan.
func folderPlan(st Status, env Env) ToolPlan {
	dir := st.OffPath
	plan := ToolPlan{Status: st, Default: "install:" + SmithPathManager}
	plan.Options = append(plan.Options, Option{
		Kind: OptionInstall, Value: plan.Default, Manager: SmithPathManager,
		Command: []string{"smith", "tools", "use", st.Name},
		Label:   words.Say("toolfolder.option.smith.label"),
		Hint:    words.Say("toolfolder.option.smith.hint", "folder", dir, "file", env.ToolFolders),
	})
	if env.OS == "windows" {
		plan.Options = append(plan.Options, Option{
			Kind: OptionInstall, Value: "install:" + UserPathManager, Manager: UserPathManager,
			Command: UserPathCommand(dir),
			Label:   words.Say("toolfolder.option.userpath.label"),
			Hint:    words.Say("toolfolder.option.userpath.hint", "folder", dir),
		})
	}
	return withEnd(plan,
		Option{Kind: OptionMyself, Value: "myself", Label: words.Say("toolfolder.option.show.label"), Hint: words.Say("toolfolder.option.show.hint")})
}

// withEnd appends the given last options and "Skip it" to a plan.
func withEnd(plan ToolPlan, last ...Option) ToolPlan {
	plan.Options = append(append(plan.Options, last...),
		Option{Kind: OptionSkip, Value: "skip", Label: "Skip it", Hint: "go on without it; " + plan.Consequence})
	return plan
}

// RenderFolder is what "Show me the folder" prints: the folder, the command
// that adds it to the user PATH, and the smith command that uses it.
func RenderFolder(plan ToolPlan) string {
	return words.Say("toolfolder.show", "title", plan.Title, "folder", plan.OffPath,
		"command", UserPathScript(plan.OffPath), "tool", plan.Name) + "\n"
}
