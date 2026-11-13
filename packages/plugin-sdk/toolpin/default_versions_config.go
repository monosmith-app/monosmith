package toolpin

// defaults is the version smith installs when the workspace pins none, by
// toolchain label. Bump a value deliberately: it moves every new pipeline,
// dev container and project scaffold.
var defaults = map[string]string{
	"node":   "24",
	"go":     "1.24",
	"python": "3.12",
	"rustc":  "1.90",
	"java":   "21",
	"dotnet": "10.0",
	"ruby":   "3.4",
	"php":    "8.4",
}

// Default is the version smith uses for a toolchain label when the workspace
// pins none. It is empty for a label smith has no default for.
func Default(label string) string {
	return defaults[label]
}
