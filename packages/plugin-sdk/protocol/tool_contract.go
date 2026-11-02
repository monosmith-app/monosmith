package protocol

// ToolNeed is a program on the person's machine that a plugin needs to build,
// test or run its projects (Python, the .NET SDK, Docker). A plugin publishes
// the tools it needs in InitializeResult.Tools. The host detects them, and when
// one is missing it presents the ways to get it (ADR 0076): it never installs
// in silence. The plugin writes facts only (what the tool is, how to detect it,
// how each system gets it); the host decides how to ask.
type ToolNeed struct {
	// Name identifies the tool across plugins and in commands (smith tools
	// install python). Two plugins that need the same tool use the same Name.
	Name string `json:"name"`
	// Title is what the tool is called in text: "Python", "the .NET SDK".
	// Doctor words the check "<Title> is installed".
	Title string `json:"title"`
	// Why says in plain words what smith needs the tool for.
	Why string `json:"why"`
	// Consequence says what stops working while the tool is missing, as the
	// end of a sentence that starts "If you skip it, ": "smith test and
	// smith build fail for the python projects".
	Consequence string `json:"consequence"`
	// Program is the program that carries the tool on PATH ("python3").
	Program string `json:"program"`
	// Alternatives are other programs that also count ("python", "py").
	Alternatives []string `json:"alternatives,omitempty"`
	// VersionArgs are the arguments that make Program print its version.
	// Absent means --version.
	VersionArgs []string `json:"versionArgs,omitempty"`
	// Pins are the files where a workspace pins the version it wants. The
	// first pin the host finds gives the version range. Absent means any
	// version counts.
	Pins []ToolVersionPin `json:"pins,omitempty"`
	// Choice limits the tool to stacks that have this choice ("azure-functions"):
	// the tool is not needed by a project built without it. Absent means every
	// project of the plugin needs it.
	Choice string `json:"choice,omitempty"`
	// Suggested is the version smith installs when no pin says one: "3.13".
	// It fills {major} and {major.minor} in the install entries.
	Suggested string `json:"suggested,omitempty"`
	// Optional marks a tool whose absence only costs a feature (running a
	// function app locally). Doctor reports it as a warning, not a failure.
	Optional bool `json:"optional,omitempty"`
	// Install are the ways to get the tool, per package manager and system.
	Install []ToolInstall `json:"install,omitempty"`
	// Download is the official page of the tool, for the person who installs
	// it by hand or has no package manager.
	Download string `json:"download"`
	// Source is the address of the page the Install entries were checked
	// against.
	Source string `json:"source"`
	// Checked is the date (YYYY-MM-DD) the entries were last checked.
	Checked string `json:"checked"`
}

// ToolVersionPin names a file that pins the version of a tool and how to read it.
// Format is one of: "text" (the first word of the file, as in .nvmrc and
// .python-version), "go-mod" (the toolchain line, else the go line),
// "pyproject" (requires-python), "global-json" (sdk.version) and
// "rust-toolchain" (the channel).
type ToolVersionPin struct {
	// File is the base name of the file, looked up at the workspace root and
	// in the folders of the projects that use the tool.
	File   string `json:"file"`
	Format string `json:"format"`
}

// ToolInstall is one way to get a tool with a package manager. Manager is one
// of: "winget", "brew", "brew-cask", "apt", "dnf", "pacman", "npm", "uv",
// "fnm" and "rustup". The host builds the exact command from the manager and
// Package; Package may hold {major} and {major.minor}, which the pinned or
// suggested version fills. The language's own installers (uv, fnm, rustup)
// are offered only when they are already on the machine.
type ToolInstall struct {
	Manager string `json:"manager"`
	Package string `json:"package"`
	// Note adds a sentence the person needs before choosing ("a new terminal
	// is needed").
	Note string `json:"note,omitempty"`
}
