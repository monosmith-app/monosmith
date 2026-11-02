package functions

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

const language = "python"

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

func runtimeChoice(name, description string, suggested bool) protocol.Choice {
	return protocol.Choice{Axis: "runtime", Name: name, Description: description, Language: language, Suggested: suggested}
}

// Choices are the choices the Python plugin publishes for function apps
// (InitializeResult.Choices): three hosts, four triggers and three runtimes.
// The runtimes are the Python versions that Azure Functions, AWS Lambda and
// Cloud Run functions all support on 2026-11-02 (ADR 0072).
func Choices() []protocol.Choice {
	return []protocol.Choice{
		hostChoice("azure-functions", "an Azure Functions app (Python v2 model): function_app.py, host.json, requirements.txt and local.settings.sample.json; runs with Core Tools (func), deploys with smith pipeline deploy-azure-function"),
		hostChoice("aws-lambda", "an AWS Lambda function: template.yaml and events/ for sam local, and a lambda_handler; runs with the SAM CLI (sam)"),
		hostChoice("gcp-functions", "a Google Cloud (Cloud Run) function: main.py and requirements.txt; runs with the Functions Framework, a dependency of the app"),
		triggerChoice("http-trigger", "answers an HTTP request (Azure: an HTTP trigger, AWS: API Gateway, Google: an HTTP function)", true),
		triggerChoice("timer-trigger", "runs on a schedule (Azure: a timer trigger, AWS: an EventBridge schedule); Google has none, use an HTTP function that Cloud Scheduler calls", false, "gcp-functions"),
		triggerChoice("queue-trigger", "answers a queue message (Azure: a Storage queue, AWS: SQS, Google: Pub/Sub)", false),
		triggerChoice("blob-trigger", "answers a new file in storage (Azure: a blob, AWS: an S3 event, Google: a Cloud Storage event)", false),
		runtimeChoice("python313", "Python 3.13; all three hosts support it", true),
		runtimeChoice("python312", "Python 3.12, the oldest all three hosts support; the workspace venv and CI use it", false),
		runtimeChoice("python314", "Python 3.14, the newest all three hosts support", false),
	}
}
