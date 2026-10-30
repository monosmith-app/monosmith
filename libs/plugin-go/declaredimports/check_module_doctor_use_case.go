package declaredimports

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// shownIssues caps how many disagreements a finding lists.
const shownIssues = 5

// Doctor is doctor's view of each module's deps-check task. It compares
// go.mod with the imports without running the go command, so doctor stays
// fast and offline; smith run deps-check asks `go mod tidy -diff` when it can.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	findings := []protocol.Finding{}
	for _, p := range params.Projects {
		f := protocol.Finding{
			Check:  p.Root + "/go.mod matches what the module imports",
			Remedy: "run `go mod tidy` in " + p.Root + " (a requirement for a module in go.work is added by hand: tidy cannot see go.work)",
		}
		result, _, err := Compare(workspaceRoot, p.Root)
		if err != nil {
			f.Detail = err.Error()
			findings = append(findings, f)
			continue
		}
		f.OK = result.Clean()
		lines := make([]string, 0, shownIssues+1)
		for i, issue := range result.Issues {
			if i == shownIssues {
				lines = append(lines, fmt.Sprintf("and %d more (smith run deps-check -p %s)", len(result.Issues)-shownIssues, p.Name))
				break
			}
			lines = append(lines, issue.String())
		}
		f.Detail = strings.Join(lines, "; ")
		findings = append(findings, f)
	}
	return protocol.DoctorResult{Findings: findings}
}
