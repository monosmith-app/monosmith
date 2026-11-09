package functions

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
)

// Data is what the templates of a function app read, next to the fields of the
// Python project: the plan and the words made from it.
type Data struct {
	Plan
	// BuildDir is the workspace's build folder of the app, relative to the app.
	BuildDir string
	// Handler is the name of the function the host calls.
	Handler string
	// HostLabel, TriggerLabel and LocalTool are words for the readme.
	HostLabel, TriggerLabel, LocalTool string
	// RuntimeID is the Python version without the dot: 313.
	RuntimeID string
}

// DataOf makes the template data of a plan; buildDir is the app's build
// folder relative to the app (protocol.BuildDirFromProject).
func DataOf(plan Plan, buildDir string) Data {
	d := Data{Plan: plan, BuildDir: buildDir, RuntimeID: strings.ReplaceAll(plan.PythonVersion, ".", "")}
	switch plan.Host {
	case functionhost.Azure:
		d.HostLabel, d.LocalTool = "Azure Functions", "Azure Functions Core Tools (func)"
		d.Handler = map[string]string{"http": "hello", "timer": "tick", "queue": "process_message", "blob": "process_blob"}[plan.Trigger]
	case functionhost.AWS:
		d.HostLabel, d.LocalTool, d.Handler = "AWS Lambda", "the SAM CLI (sam)", "lambda_handler"
	default:
		d.HostLabel, d.LocalTool = "Google Cloud (Cloud Run) functions", "the Functions Framework"
		d.Handler = map[string]string{"http": "hello", "queue": "process_message", "blob": "process_object"}[plan.Trigger]
	}
	d.TriggerLabel = map[string]string{"http": "an HTTP request", "timer": "a schedule", "queue": "a queue message", "blob": "a new file in storage"}[plan.Trigger]
	return d
}
