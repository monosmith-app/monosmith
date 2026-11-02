package functions

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

const language = "dotnet"

// hostChoice is one host: it writes the project and needs a trigger and a
// runtime, which the person answers.
func hostChoice(name, description string) protocol.Choice {
	return protocol.Choice{
		Axis: "framework", Name: name, Description: description, Language: language,
		Roles: []protocol.Role{protocol.RoleScaffold}, Needs: []string{"trigger", "runtime"},
	}
}

func triggerChoice(name, description string, suggested bool, conflicts ...string) protocol.Choice {
	return protocol.Choice{Axis: "trigger", Name: name, Description: description, Language: language, Suggested: suggested, Conflicts: conflicts}
}

func runtimeChoice(name, description string, suggested bool, conflicts ...string) protocol.Choice {
	return protocol.Choice{Axis: "runtime", Name: name, Description: description, Language: language, Suggested: suggested, Conflicts: conflicts}
}

// Choices are the choices the .NET plugin publishes for function apps
// (InitializeResult.Choices): three hosts, four triggers and two runtimes. The
// runtimes are the LTS versions of .NET that the hosts support on 2026-11-02
// (ADR 0072); .NET 8 is not offered for Google, which retires it on 2026-11-10.
func Choices() []protocol.Choice {
	return []protocol.Choice{
		hostChoice("azure-functions", "an Azure Functions app (C# isolated worker): Program.cs, host.json, local.settings.sample.json and a csproj with the Worker packages; runs with Core Tools (func), deploys with smith pipeline deploy-azure-function"),
		hostChoice("aws-lambda", "an AWS Lambda function (C# class library): template.yaml and Events/ for sam local, and a handler with the Lambda serializer; runs with the SAM CLI (sam)"),
		hostChoice("gcp-functions", "a Google Cloud (Cloud Run) function (C#): a Functions Framework class and a csproj with Google.Cloud.Functions.Hosting; runs with dotnet run, no extra tool"),
		triggerChoice("http-trigger", "answers an HTTP request (Azure: an HTTP trigger, AWS: API Gateway, Google: an HTTP function)", true),
		triggerChoice("timer-trigger", "runs on a schedule (Azure: a timer trigger, AWS: an EventBridge schedule); Google has none, use an HTTP function that Cloud Scheduler calls", false, "gcp-functions"),
		triggerChoice("queue-trigger", "answers a queue message (Azure: a Storage queue, AWS: SQS, Google: Pub/Sub)", false),
		triggerChoice("blob-trigger", "answers a new file in storage (Azure: a blob, AWS: an S3 event, Google: a Cloud Storage event)", false),
		runtimeChoice("dotnet10", ".NET 10, the current LTS; all three hosts support it", true),
		runtimeChoice("dotnet8", ".NET 8, the previous LTS; Azure and AWS support it until November 2026", false, "gcp-functions"),
	}
}
