package functions

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
)

// DetectHost says which function host a .NET project is written for: the Azure
// Functions worker package with a host.json, a SAM template.yaml for AWS Lambda,
// or the Google Functions Framework package. It returns "" for any other
// project.
//
// Parameters: dir is the absolute folder of the project; csproj is the text of
// its project file.
func DetectHost(dir, csproj string) string {
	template, _ := os.ReadFile(filepath.Join(dir, "template.yaml"))
	_, hostJSON := os.Stat(filepath.Join(dir, "host.json"))
	switch {
	case strings.Contains(csproj, "Microsoft.Azure.Functions.Worker") && hostJSON == nil:
		return functionhost.Azure
	case strings.Contains(string(template), "AWS::Serverless-2016-10-31"):
		return functionhost.AWS
	case strings.Contains(csproj, "Google.Cloud.Functions."):
		return functionhost.GCP
	}
	return ""
}

// ProjectFile reads the csproj in a project folder; "" when there is none.
func ProjectFile(dir string) string {
	matches, _ := filepath.Glob(filepath.Join(dir, "*.csproj"))
	if len(matches) == 0 {
		return ""
	}
	raw, _ := os.ReadFile(matches[0])
	return string(raw)
}
