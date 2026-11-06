package functions

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
)

// Package versions the function projects pin, checked on 2026-11-02; bump them
// deliberately.
const (
	workerVersion        = "2.52.0"
	workerSdkVersion     = "2.1.0"
	httpAspNetCoreVer    = "2.1.1"
	timerVersion         = "4.3.1"
	queuesVersion        = "5.5.5"
	blobsVersion         = "6.8.2"
	lambdaCoreVersion    = "3.4.0"
	lambdaJsonVersion    = "3.0.1"
	lambdaGatewayVersion = "3.0.1"
	lambdaSQSVersion     = "3.0.1"
	lambdaS3Version      = "4.0.0"
	lambdaEventsVersion  = "5.1.0"
	lambdaTestVersion    = "4.3.0"
	gcpHostingVersion    = "3.0.1"
	gcpEventsVersion     = "1.8.0"
)

// Data is what the templates of a function app read, next to the fields of the
// .NET project: the plan and the words and paths made from it.
type Data struct {
	Plan
	// Up leads from the app folder to the workspace root: ../../.
	Up string
	// PublishDir is the folder dotnet publish writes to, relative to the app
	// (Directory.Build.props sends every build to the workspace's dist/dotnet).
	PublishDir string
	// HostLabel and LocalTool are words for the readme.
	HostLabel, LocalTool string
	// TriggerLabel says what the function answers.
	TriggerLabel string
}

// PublishFolder is where dotnet publish -c Release puts an app, relative to the
// workspace root: the artifacts layout of Directory.Build.props.
func PublishFolder(projectID string) string {
	return "dist/dotnet/publish/" + projectID + "/release"
}

// DataOf makes the template data of a plan for an app in dir (workspace
// relative) with the project id.
func DataOf(plan Plan, dir, projectID string) Data {
	up := strings.Repeat("../", strings.Count(strings.Trim(dir, "/"), "/")+1)
	d := Data{Plan: plan, Up: up, PublishDir: up + PublishFolder(projectID)}
	switch plan.Host {
	case functionhost.Azure:
		d.HostLabel, d.LocalTool = "Azure Functions", "Azure Functions Core Tools (func)"
	case functionhost.AWS:
		d.HostLabel, d.LocalTool = "AWS Lambda", "the SAM CLI (sam)"
	default:
		d.HostLabel, d.LocalTool = "Google Cloud (Cloud Run) functions", "dotnet run"
	}
	d.TriggerLabel = map[string]string{"http": "an HTTP request", "timer": "a schedule", "queue": "a queue message", "blob": "a new file in storage"}[plan.Trigger]
	return d
}
