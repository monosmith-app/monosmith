package venv

// PythonVersion is the interpreter version generated pipelines install.
const PythonVersion = "3.12"

// RequirementsFile is the workspace's development toolchain, nothing else.
// Runtime dependencies belong to each project's pyproject.toml.
const RequirementsFile = "requirements-dev.txt"

// Requirements pins the tools; bump them deliberately.
const Requirements = `# The workspace's Python development toolchain, installed into .venv.
# Runtime dependencies belong in each project's pyproject.toml, never here.
build==1.6.1
mypy==2.4.0
pip-audit==2.10.1
pytest==9.1.1
ruff==0.17.0
setuptools==84.2.1
twine==7.0.0
`
