package functions

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
)

// Plan is what the choices of a stack say about a function app.
type Plan struct {
	// Host is azure-functions, aws-lambda or gcp-functions.
	Host string
	// Trigger is http, timer, queue or blob.
	Trigger string
	// PythonVersion is the runtime, as "3.13".
	PythonVersion string
}

// runtimes maps the runtime choices to their Python versions.
var runtimes = map[string]string{"python312": "3.12", "python313": "3.13", "python314": "3.14"}

// PlanOf reads the choices of a new project. The second result is false when
// none of them is a function host: the project is an ordinary Python app.
func PlanOf(choices []string) (Plan, bool) {
	var plan Plan
	for _, c := range choices {
		switch {
		case slices.Contains([]string{functionhost.Azure, functionhost.AWS, functionhost.GCP}, c):
			plan.Host = c
		case strings.HasSuffix(c, "-trigger"):
			plan.Trigger = strings.TrimSuffix(c, "-trigger")
		case runtimes[c] != "":
			plan.PythonVersion = runtimes[c]
		}
	}
	return plan, plan.Host != ""
}
