package pipeline

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-python/registries"
	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CI returns the Python contribution. Setup steps are computed on the machine
// that runs them, so the interpreter path matches that OS.
func CI(params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	py := venv.Interpreter()
	launcher := venv.Launchers()[0]
	setup := []protocol.Step{
		{Name: "Create the Python venv", Command: append(append([]string{}, launcher...), "-m", "venv", "--upgrade-deps", ".venv"), Blocking: true},
	}
	// The interpreter's version comes first, so a version that differs between
	// runners (a resolver failure on one OS only) is visible in the log.
	setup = append(setup, protocol.Step{Name: "Show the Python version", Command: []string{py, "--version"}})
	// A private index goes into the venv's own pip configuration, so every
	// install through the venv uses it. On Azure Pipelines PipAuthenticate@1
	// sets PIP_INDEX_URL with the pipeline's token, which takes precedence.
	index := registries.Index(params.Projects)
	if index != "" {
		setup = append(setup, protocol.Step{Name: "Point pip at the private index", Command: []string{py, "-m", "pip", "config", "--site", "set", "global.index-url", index}, Blocking: true})
	}
	setup = append(setup, protocol.Step{Name: "Install the Python toolchain", Command: []string{py, "-m", "pip", "install", "--quiet", "--disable-pip-version-check", "-r", venv.RequirementsFile}, Blocking: true})
	// Project installs run without --quiet: it hides the resolver's report of
	// what conflicts, the one thing a failed install needs to show.
	// A library comes before the project that depends on it, so pip finds it in
	// the venv and never looks its name up on PyPI.
	for _, p := range venv.DependencyOrder(params.Projects) {
		setup = append(setup, protocol.Step{Name: "Install " + p.Name, Command: []string{py, "-m", "pip", "install", "--disable-pip-version-check", "-e", "./" + p.Root}, Blocking: true})
	}
	result := protocol.CIResult{
		Toolchains: []protocol.Toolchain{runtime(venv.PythonVersion)},
		Setup:      setup,
		Stores:     []protocol.PackageStore{{Name: "pip", Path: "~/.cache/pip", Env: "PIP_CACHE_DIR", Files: dependencyFiles(params.Projects)}},
		Audit:      []protocol.Step{{Name: "pip-audit", Command: []string{py, "-m", "pip_audit", "--skip-editable"}, Blocking: false, CacheFiles: dependencyFiles(params.Projects)}},
	}
	if params.Audit.Enabled("python") {
		// Opt-in ([ci] audit_policies): hashed pins, no direct URLs, only the
		// workspace's index, in the toolchain file and each project's
		// requirements.txt.
		files := []string{venv.RequirementsFile}
		for _, p := range params.Projects {
			files = append(files, path.Join(p.Root, "requirements.txt"))
		}
		command := []string{"{smith}", "py-policy"}
		if index != "" {
			command = append(command, "--index="+index)
		}
		result.Audit = append(result.Audit, protocol.Step{Name: "python policy (requirements)", Command: append(command, files...), Blocking: params.Audit.Blocks("python"), CacheFiles: files})
	}
	signIn, publicPublish := registries.CI(params.Projects)
	result.Authenticate = signIn.Authenticate
	result.Credentials = append(result.Credentials, signIn.Credentials...)
	for _, p := range params.Projects {
		if p.Publishes() && publicPublish {
			result.Credentials = append(result.Credentials, protocol.Credential{Env: "PYPI_TOKEN", Description: "PyPI API token (or set TWINE_USERNAME/TWINE_PASSWORD and TWINE_REPOSITORY_URL)", Required: true})
			break
		}
	}
	return result
}

// Doctor checks the venv and the requirements file exist, and that the venv's
// Python is the one CI installs.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	_, venvErr := os.Stat(venv.Path(workspaceRoot))
	_, reqErr := os.Stat(filepath.Join(workspaceRoot, venv.RequirementsFile))
	findings := []protocol.Finding{
		{Check: "Python venv exists", OK: venvErr == nil, Detail: ".venv is missing at the workspace root", Remedy: "run `smith ci setup` (or `python -m venv .venv` and install requirements-dev.txt)"},
		{Check: "requirements-dev.txt exists", OK: reqErr == nil, Detail: "the Python toolchain file is missing", Remedy: "run `smith add` for a Python project, or restore requirements-dev.txt"},
	}
	if venvErr == nil {
		findings = append(findings, CheckPythonVersion(workspaceRoot))
	}
	if f, ok := CheckVenvIgnored(workspaceRoot); ok {
		findings = append(findings, f)
	}
	for _, p := range params.Projects {
		if p.Kind == protocol.ProjectPackage {
			findings = append(findings, CheckPackageImports(workspaceRoot, p))
		}
	}
	return protocol.DoctorResult{Findings: findings}
}

// dependencyFiles are the files that decide what the venv holds, and so what
// pip-audit finds in it and what pip downloads: the toolchain's requirements
// and each project's pyproject.toml.
func dependencyFiles(projects []protocol.Project) []string {
	files := []string{venv.RequirementsFile}
	for _, p := range projects {
		files = append(files, path.Join(p.Root, "pyproject.toml"))
	}
	return files
}
