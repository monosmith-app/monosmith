package protocol

// DoctorParams gives the plugin its projects to check.
type DoctorParams struct {
	Projects []Project `json:"projects"`
}

// DoctorResult carries the plugin's findings.
type DoctorResult struct {
	Findings []Finding `json:"findings"`
}

// Finding is one check. Every failed finding names its remedy.
type Finding struct {
	Check  string `json:"check"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
	Remedy string `json:"remedy,omitempty"`
	// Warning true reports a problem without failing doctor.
	Warning bool `json:"warning,omitempty"`
	// Fix is the change that clears a failed finding, when the plugin can
	// write it; smith doctor --fix shows it as a diff before applying it.
	Fix *Fix `json:"fix,omitempty"`
}

// Fix is a change that clears a failed finding: files' new contents, then
// commands to run, and questions the user answers first. Every part is
// optional; the host shows all of it before applying anything.
//
// A host that does not know Steps or Choices ignores them, so a plugin marks
// a fix Safe only when its Files alone are sound.
type Fix struct {
	Description string `json:"description"`
	// Safe marks a fix that only restores what smith or the plugin owns and
	// loses nothing of the user's; doctor --fix --yes applies only safe fixes
	// with no Choices. Any other fix is asked about before it runs.
	Safe bool `json:"safe"`
	// Files are written first, as whole new contents.
	Files []FileEdit `json:"files,omitempty"`
	// Steps run after the files, in order, from the workspace root; the first
	// one that fails stops the rest.
	Steps []FixStep `json:"steps,omitempty"`
	// Choices are questions the user answers before the fix applies, in order.
	// The answer's Fix is added to this one: its files and steps come after
	// this fix's own. A fix with Choices is never applied unasked.
	Choices []FixChoice `json:"choices,omitempty"`
	// AppliesTo names the projects the fix is for. The host folds fixes that
	// share a Description and name projects into one ("applies to 22
	// projects"), so give the same Description to the same kind of fix on
	// every project. A fix that edits one shared file for many projects lists
	// them all here.
	AppliesTo []string `json:"applies_to,omitempty"`
}

// FixStep is a command to run in the workspace.
type FixStep struct {
	// Description says in words what the command does.
	Description string `json:"description"`
	// Argv is the program and its arguments; no shell is involved.
	Argv []string `json:"argv"`
	// Dir is where it runs, workspace-relative; empty is the workspace root.
	Dir string `json:"dir,omitempty"`
}

// FixChoice is a question with options, one wizard step.
type FixChoice struct {
	Question string            `json:"question"`
	Options  []FixChoiceOption `json:"options"`
}

// FixChoiceOption is one answer: a label, and the Fix it adds when picked. The
// first option is the suggested one.
type FixChoiceOption struct {
	Label string `json:"label"`
	Fix   Fix    `json:"fix"`
}

// FileEdit is one file's new content, workspace-relative; Delete removes it.
type FileEdit struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
	Delete  bool   `json:"delete,omitempty"`
}
