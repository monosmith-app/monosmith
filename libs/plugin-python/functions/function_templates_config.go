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

import json
import re
import shutil
import sys
import tomllib
from importlib.metadata import PackageNotFoundError, distribution
from pathlib import Path
from urllib.parse import urlparse
from urllib.request import url2pathname

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


def requirement_name(requirement: str) -> str:
    """The distribution name of a requirement line ("core-py>=1" gives "core-py")."""
    return re.split(r"[\s;<>=!~\[(]", requirement.strip(), maxsplit=1)[0]


def workspace_library(name: str) -> Path | None:
    """The folder of a library of this workspace, or None for a package of a registry.

    smith installs a library of the workspace into the venv from its folder, and pip
    records that folder in direct_url.json. The registry has no such package, so the
    zip must carry the library's code.
    """
    try:
        raw = distribution(name).read_text("direct_url.json")
    except PackageNotFoundError:
        return None
    url = urlparse(json.loads(raw).get("url", "")) if raw else None
    return Path(url2pathname(url.path)) if url and url.scheme == "file" else None


def library_packages(folder: Path, name: str) -> list[Path]:
    """The importable packages of a library: its import name under src/, else every package there."""
    roots = [folder / "src", folder]
    wanted = name.replace("-", "_").lower()
    for root in roots:
        if (root / wanted / "__init__.py").is_file():
            return [root / wanted]
    return [p.parent for p in sorted((folder / "src").glob("*/__init__.py"))]


def workspace_packages() -> list[Path]:
    """The packages of the workspace libraries the app depends on, and of the libraries they depend on."""
    packages: dict[str, list[Path]] = {}
    pending = tomllib.loads(Path("pyproject.toml").read_text(encoding="utf-8")).get("project", {}).get("dependencies", [])
    while pending:
        name = requirement_name(pending.pop())
        folder = None if name in packages else workspace_library(name)
        if folder is None:
            continue
        packages[name] = library_packages(folder, name)
        if not packages[name]:
            sys.exit(f"{name} is a library of this workspace but has no package to ship: look in {folder}")
        pending.extend(distribution(name).requires or [])
    return [p for found in packages.values() for p in found]


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
    # The libraries of the workspace are not on the registry: the code goes into the zip.
    for package in workspace_packages():
        shutil.copytree(package, STAGE / package.name, ignore=IGNORE, dirs_exist_ok=True)


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
