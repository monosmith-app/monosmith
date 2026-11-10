package agentdoc

// The generated text of a file sits between these two lines. A file of the
// same region name in another tool (see ADR 0046) uses the same form.
const (
	// BeginMarker opens the generated region.
	BeginMarker = "<!-- smith:begin agent-instructions -->"
	// EndMarker closes the generated region.
	EndMarker = "<!-- smith:end agent-instructions -->"
)

// File names the command writes, relative to the workspace root.
const (
	// AgentsFile holds the instructions.
	AgentsFile = "AGENTS.md"
	// ClaudeFile points Claude Code to AgentsFile.
	ClaudeFile = "CLAUDE.md"
)

// namingByLanguage is how a language spells a file name under the
// vertical-slices convention.
var namingByLanguage = map[string]string{
	"go":     "`<name>_<role>.go`, tests `<name>_<role>_test.go`",
	"python": "`<name>_<role>.py`, tests `test_<name>.py`",
	"node":   "`<name>.<role>.ts` with the role in kebab-case (`place-order.use-case.ts`), checked by the ESLint config",
	"rust":   "`<name>_<role>.rs` in a folder per slice",
	"ruby":   "`<name>_<role>.rb` in a folder per slice",
	"dart":   "`<name>_<role>.dart`, lowercase with underscores",
	"cpp":    "`<name>_<role>` or `<Name><Role>`, one spelling for a folder",
	"dotnet": "`<Name><Role>.cs` in PascalCase, one folder per slice",
	"jvm":    "`<Name><Role>.kt` or `.java` in PascalCase, one package per slice",
	"php":    "`<Name><Role>.php` in PascalCase, one folder per slice",
}
