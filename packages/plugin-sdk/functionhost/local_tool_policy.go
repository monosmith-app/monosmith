package functionhost

// The names of the function-host choices (protocol.Choice.Name).
const (
	Azure = "azure-functions"
	AWS   = "aws-lambda"
	GCP   = "gcp-functions"
)

// Tool is the program a person runs a host's functions locally with.
type Tool struct {
	// Host is the choice name of the host.
	Host string
	// Name is what the tool is called in text.
	Name string
	// Binary is the program on PATH; empty when the tool is a dependency of the
	// project, which the language plugin checks in its own environment.
	Binary string
	// Install says how to get it, in one sentence.
	Install string
}

// tools are the local tools by host.
var tools = map[string]Tool{
	Azure: {
		Host: Azure, Name: "Azure Functions Core Tools", Binary: "func",
		Install: "npm install -g azure-functions-core-tools@4 (or brew tap azure/functions && brew install azure-functions-core-tools@4, or winget install Microsoft.Azure.FunctionsCoreTools); smith pipeline deploy-azure-function deploys the app",
	},
	AWS: {
		Host: AWS, Name: "the AWS SAM CLI", Binary: "sam",
		Install: "brew install aws-sam-cli, winget install Amazon.SAM-CLI or pip install aws-sam-cli (https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html); sam local also needs Docker",
	},
	GCP: {
		Host: GCP, Name: "the Functions Framework", Binary: "",
		Install: "install the project's dependencies: smith install",
	},
}

// ToolOf returns the local tool of a host choice; false for any other name.
func ToolOf(host string) (Tool, bool) {
	t, ok := tools[host]
	return t, ok
}
