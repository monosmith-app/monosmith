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
// file, the packaging script and the readme.
var commonTemplates = map[string]string{
	"[[.Dir]]/src/[[.Import]]/functions/__init__.py": `[[- if eq .Host "azure-functions"]]"""The function of the app: its blueprint for the Azure Functions host."""

from .greeting_handler import blueprint, [[.Handler]]

__all__ = ["blueprint", "[[.Handler]]"]
[[- else]]"""The function of the app."""

from .greeting_handler import [[.Handler]]

__all__ = ["[[.Handler]]"]
[[- end]]
`,
	"[[.Dir]]/package.py": `"""Stages the deployed files into the workspace's dist/ ("build") and zips them ("zip")."""

import shutil
import sys
from pathlib import Path

OUT = Path("[[.BuildDir]]")
STAGE = OUT / "app"
[[- if eq .Host "aws-lambda"]]
# Lambda loads the package from the root of the zip: src/ holds the package.
SOURCES = ["src"]
FLATTEN = True
[[- else if eq .Host "azure-functions"]]
SOURCES = ["function_app.py", "host.json", "requirements.txt", "src"]
FLATTEN = False
[[- else]]
SOURCES = ["main.py", "requirements.txt", "src"]
FLATTEN = False
[[- end]]
IGNORE = shutil.ignore_patterns("__pycache__", "*.egg-info", "*.pyc")


def build() -> None:
    """Copy the files that are deployed into the build folder."""
    shutil.rmtree(STAGE, ignore_errors=True)
    STAGE.mkdir(parents=True)
    for source in SOURCES:
        path = Path(source)
        if path.is_dir():
            target = STAGE if FLATTEN else STAGE / path
            shutil.copytree(path, target, ignore=IGNORE, dirs_exist_ok=True)
        else:
            shutil.copy2(path, STAGE / path)


def pack() -> None:
    """Pack the build folder into [[.Name]].zip next to it."""
    archive = shutil.make_archive(str(OUT / "[[.Name]]"), "zip", STAGE)
    print(archive)


if __name__ == "__main__":
    {"build": build, "zip": pack}[sys.argv[1]]()
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
