package toolinstall

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// managedValue is the Value of the option that installs a tool into smith's
// own folder; --manager go-managed picks it.
const managedValue = "install:go-managed"

// withManagedCopy puts first, as the recommended answer, the install of a tool
// with one wanted version into smith's own folder for the workspace
// (.smith/bin). The lint task and the doctor look there before PATH, so the
// pinned copy is the one that runs even when another copy comes first on PATH.
// The other copy is left alone and PATH is not edited. It needs the go
// install option, so it exists only where go is on the machine.
func withManagedCopy(options []Option, st Status, version string) []Option {
	if !st.Exact || len(st.Dirs) == 0 {
		return options
	}
	i := slices.IndexFunc(options, func(o Option) bool { return o.Value == "install:go" })
	if i < 0 {
		return options
	}
	dir := toolrun.ManagedBinDir(st.Dirs[0])
	managed := options[i]
	managed.Value, managed.Manager = managedValue, "go-managed"
	managed.Env = []string{"GOBIN=" + dir}
	managed.Note = "go install compiles it into " + dir + "; nothing else on this machine changes"
	managed.Label = "Install " + version + " into smith's own folder (recommended)"
	managed.Hint = "runs " + strings.Join(managed.Env, " ") + " " + strings.Join(managed.Command, " ") +
		"; smith's lint and doctor use that copy first, and your other copy and PATH stay as they are"
	options[i].Hint += "; the copy lands in the Go bin folder and may stay behind another copy on PATH"

	return append([]Option{managed}, options...)
}
