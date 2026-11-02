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
	// Runtime is the choice name of the .NET version: dotnet10 or dotnet8.
	Runtime string
	// Framework is the target framework moniker of the runtime: net10.0.
	Framework string
}

// frameworks maps the runtime choices to their target framework monikers.
var frameworks = map[string]string{"dotnet8": "net8.0", "dotnet10": "net10.0"}

// PlanOf reads the choices of a new project. The second result is false when
// none of them is a function host: the project is an ordinary .NET app.
func PlanOf(choices []string) (Plan, bool) {
	var plan Plan
	for _, c := range choices {
		switch {
		case slices.Contains([]string{functionhost.Azure, functionhost.AWS, functionhost.GCP}, c):
			plan.Host = c
		case strings.HasSuffix(c, "-trigger"):
			plan.Trigger = strings.TrimSuffix(c, "-trigger")
		case frameworks[c] != "":
			plan.Runtime, plan.Framework = c, frameworks[c]
		}
	}
	return plan, plan.Host != ""
}
