package functions

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
)

// DetectHost says which function host a Python project is written for, by the
// files the host needs: function_app.py with host.json for Azure Functions, a
// SAM template.yaml for AWS Lambda, main.py with the Functions Framework for
// Google. It returns "" for any other project.
//
// Parameters: dir is the absolute folder of the project.
func DetectHost(dir string) string {
	has := func(name string) bool {
		_, err := os.Stat(filepath.Join(dir, name))
		return err == nil
	}
	read := func(name string) string {
		raw, _ := os.ReadFile(filepath.Join(dir, name))
		return string(raw)
	}
	switch {
	case has("function_app.py") && has("host.json"):
		return functionhost.Azure
	case strings.Contains(read("template.yaml"), "AWS::Serverless-2016-10-31"):
		return functionhost.AWS
	case has("main.py") && strings.Contains(read("requirements.txt"), "functions-framework"):
		return functionhost.GCP
	}
	return ""
}
