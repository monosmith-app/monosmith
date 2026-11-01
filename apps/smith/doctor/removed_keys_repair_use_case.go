package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// repairRemovedKeys is what doctor does when smith.toml uses a table smith
// renamed, so no session can open. It reads smith.toml as text, reports the
// old keys as a finding, and with fix shows the rename as a diff and applies it
// once the user confirms (or with yes). It returns true when the file was
// changed, so the caller can open the workspace and run the full checks. It
// returns ErrUnhealthy when the keys are left as they are.
func repairRemovedKeys(cmd *cobra.Command, asJSON, fix, yes bool) (bool, error) {
	root, err := workspace.FindRoot(".")
	if err != nil {
		return false, err
	}
	raw, err := os.ReadFile(filepath.Join(root, workspace.FileName))
	if err != nil {
		return false, err
	}
	finding, ok := RemovedKeysFinding(raw)
	if !ok {
		return false, nil
	}
	out := cmd.OutOrStdout()
	report := Report{Findings: []Finding{{Finding: finding, Source: "smith"}}}
	report.tally()
	report.Fixes = Plan(report)
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		_ = enc.Encode(report)
		return false, ErrUnhealthy
	}
	printRemovedKeysFinding(cmd, finding)
	if !fix {
		fmt.Fprintln(out, "\n"+fixesAvailable(report.Fixes))
		return false, ErrUnhealthy
	}
	fmt.Fprintln(out)
	res, err := doctorfix.Walk(cmd.Context(), report.Fixes, doctorfix.Options{
		Root: root, Out: out, In: cmd.InOrStdin(), Interactive: term.IsTerminal(int(os.Stdin.Fd())), Yes: yes,
		Command: func(p doctorfix.Planned, answers []int) string { return retryCommand(p, answers) },
	})
	if err != nil {
		return false, err
	}
	if len(res.Applied) == 0 {
		return false, ErrUnhealthy
	}
	fmt.Fprintln(out, "\nchecking again with the renamed keys")
	return true, nil
}

// printRemovedKeysFinding prints the finding the way the full report does.
func printRemovedKeysFinding(cmd *cobra.Command, f protocol.Finding) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "✖ %s\n    %s\n", f.Check, f.Detail)
	for _, step := range strings.Split(f.Remedy, "\n") {
		fmt.Fprintf(out, "    → %s\n", step)
	}
}
