package adoptsplit

// The ecosystems of the project found at the repository root.
const (
	Node   = "node"
	Go     = "go"
	Python = "python"
	Rust   = "rust"
)

// Rules is what one language declares about the split of a single-project
// repository. The zero value is a language that declares nothing: every
// top-level entry moves with the project.
type Rules struct {
	// Ecosystem is one of the ecosystem constants.
	Ecosystem string
	// Workspace are the top-level names, as path.Match patterns in lower case, of
	// tool configuration that belongs to the workspace and stays at the root:
	// formatter, linter and similar files that apply to every project.
	Workspace []string
	// Lockfiles are the top-level names, as path.Match patterns in lower case,
	// that stay at the root because the workspace has one lockfile for all.
	Lockfiles []string
	// Sort reads the manifest at the repository root, before anything moves,
	// and sorts the dependencies. It is nil when the language has no
	// dependency split.
	Sort func(root string) (Sorted, error)
	// Split edits the files after the move: it takes the tools out of the
	// project's manifest, writes the workspace's, and shares the configuration.
	// It is nil when the language moves its manifest whole.
	Split func(Request) (Result, error)
}

// Sorted is the dependencies of the manifest, sorted by where they belong.
type Sorted struct {
	// Tools are the dependencies that are tools (compilers, linters, formatters,
	// test runners and their plugins and types): they go to the workspace
	// without a question.
	Tools []string
	// Lint is true when the project has a lint setup of its own, which the
	// shared tooling would replace: the person chooses between the two.
	Lint bool
	// Replaced are the dependencies the workspace's shared tooling replaces: they
	// leave the project and do not go to the workspace.
	Replaced []string
	// Unsure are the dependencies of development smith does not know: the
	// person says where each goes.
	Unsure []string
}

// Request is what a Split needs.
type Request struct {
	// Root is the repository root, as a path of the machine.
	Root string
	// Dir is the project's folder, from the root, with forward slashes.
	Dir string
	// Workspace is the repository's name, which names the workspace manifest.
	Workspace string
	// Patterns are the entries of the workspace manifest that list the projects.
	Patterns []string
	// Hoist are the dependencies that go to the workspace: the tools and the
	// ones the person chose.
	Hoist []string
	// Drop are the dependencies the shared tooling replaces.
	Drop []string
	// KeepLint is true when the person keeps the project's own lint setup: it
	// moves into the project folder and nothing of it is replaced.
	KeepLint bool
}

// Result is what a Split did.
type Result struct {
	// Notes are things the person should know, one sentence each.
	Notes []string
	// ESLint is true when the project used ESLint, so the workspace needs the
	// shared lint configuration.
	ESLint bool
}
