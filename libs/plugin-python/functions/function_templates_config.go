package functions

import (
	"maps"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
)

// Dependencies are the runtime requirements of the app, as pyproject.toml lists
// them: the library of the host. Lambda has its libraries in the runtime.
func Dependencies(host string) string {
	switch host {
	case functionhost.Azure:
		return `["azure-functions>=1.25"]`
	case functionhost.GCP:
		return `["functions-framework>=3.10"]`
	}
	return "[]"
}

// Templates are the files a function app adds to the Python app's own (the
// pyproject.toml, the greeting slice and its test), with [[ ]] delimiters. The
// data has Dir, Name, Import, Host, Trigger, PythonVersion and BuildDir.
func Templates(host string) map[string]string {
	files := maps.Clone(commonTemplates)
	switch host {
	case functionhost.Azure:
		maps.Copy(files, azureTemplates)
	case functionhost.AWS:
		maps.Copy(files, awsTemplates)
	case functionhost.GCP:
		maps.Copy(files, gcpTemplates)
	}
	return files
}

// commonTemplates are the files of every host: the functions slice's package
// file and the readme.
var commonTemplates = map[string]string{
	"[[.Dir]]/src/[[.Import]]/functions/__init__.py": `[[- if eq .Host "azure-functions"]]"""The function of the app: its blueprint for the Azure Functions host."""

from .greeting_handler import blueprint, [[.Handler]]

__all__ = ["blueprint", "[[.Handler]]"]
[[- else]]"""The function of the app."""

from .greeting_handler import [[.Handler]]

__all__ = ["[[.Handler]]"]
[[- end]]
`,
	"[[.Dir]]/README.md": `# [[.Name]]

A [[.HostLabel]] app in Python [[.PythonVersion]], made by smith. It answers [[.TriggerLabel]].

- ` + "`smith test`" + ` runs the unit tests of the function: they need no cloud tool.
- ` + "`smith run start`" + ` runs it locally with [[.LocalTool]].
- ` + "`smith run build`" + ` stages the deployed files in dist/ and ` + "`smith run package`" + ` zips them.
[[- if eq .Host "azure-functions"]]
- ` + "`smith pipeline deploy-azure-function`" + ` adds a pipeline that deploys the app to an Azure Function App.
- Copy local.settings.sample.json to local.settings.json (git ignores it) for the settings of your machine.
- Create the Function App with Python [[.PythonVersion]] as its runtime.
[[- else if eq .Host "aws-lambda"]]
- template.yaml is the SAM template (Runtime python[[.PythonVersion]]); events/greeting.json is the sample event.
[[- else]]
- Deploy it with ` + "`gcloud run deploy`" + ` or ` + "`gcloud functions deploy`" + ` and the runtime python[[.RuntimeID]].
[[- end]]
`,
}
