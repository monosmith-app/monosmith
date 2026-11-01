package why

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// explainSetting answers `smith why ci.devcontainer`: when the argument is
// the path of a setting that smith config edits ("table.key") and no file has
// that name, it says where the value comes from and why the default is what
// it is. It reports false, and does nothing, when the argument is not a
// setting, so the caller treats it as a file.
//
// Parameters: cmd prints the answer; arg is the argument; asJSON prints the
// answer as JSON.
// Returns whether it answered, and an error when the workspace cannot be read.
func explainSetting(cmd *cobra.Command, arg string, asJSON bool) (bool, error) {
	if strings.ContainsAny(arg, `/\`) || !strings.Contains(arg, ".") {
		return false, nil
	}
	if _, err := os.Stat(arg); err == nil {
		return false, nil
	}
	if _, err := configedit.Find(arg); err != nil {
		return false, nil
	}
	root, err := workspace.FindRoot(".")
	if err != nil {
		return true, err
	}
	w, err := configedit.ExplainSetting(root, arg)
	if err != nil {
		return true, err
	}
	out := cmd.OutOrStdout()
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return true, enc.Encode(w)
	}
	fmt.Fprint(out, w.Text())
	return true, nil
}
