# The python plugin

The python plugin makes every `pyproject.toml` that has a `[project]` table a smith project, and gives all of them one shared virtual environment. It creates `.venv` at the workspace root, installs every project into it in editable mode (so projects can import each other), runs Ruff, mypy and pytest through that interpreter, on smith's default configuration unless the repository has its own, builds wheels and sdists with PyPA `build`, and publishes packages to PyPI or a private index with twine. It is a built-in plugin: no installation, no setup beyond having a Python on the machine.

## At a glance

| | |
|---|---|
| Plugin name | `python` (built-in, version follows smith, 1.62.0 here) |
| Roles | `language`, `toolchain`, `scaffold` (all three by default for a language plugin) |
| `--lang` values | `python`, `py` |
| Manifest that makes a directory a project | `pyproject.toml` with a `[project]` table that has a `name` |
| Requires / conflicts / provides | none / none / no task slots |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor`, `remove`, `published`, `dependencies`, `declarations`, `latest`, `changeKind`, `components`, `channels`, `declared`, `link` |
| Stacks it publishes | none |
| Project kinds it can create | `app`, `lib`, `package` (`python-app`, `python-lib`, `python-package`) |

Capabilities are from `smith plugin list --json`.

## What it detects

(`libs/plugin-python/discovery/`)

| Field | How it is read |
|---|---|
| Project | each `pyproject.toml` below the workspace root. A root `pyproject.toml` configures tools and is not a project. A file with no `[project].name` (a tool-only file) is skipped. smith's directory layout is not required: the scan does not insist on `src/`. |
| Name | the last element of the directory (`packages/sdk` is `sdk`), not the distribution name |
| Kind | `[tool.smith] kind = "app" \| "lib" \| "package"`; else the top directory (`apps/`, `libs/`, `packages/`); else `lib` |
| Stack | `[tool.smith] stack = "..."` |
| Version | `[project] version` |
| Description | `[project] description`, trimmed |
| Exports | the import name: the distribution name lower-cased with `-` and `.` replaced by `_` (`py-demo-util` exports `py_demo_util`) |
| Imports (graph edges) | the first segment of every `import x` / `from x import y` in every `.py` file under the project (absolute imports only; relative imports are ignored), plus the import name of every `[project].dependencies` requirement. Skipped directories: `dist`, `build`, `node_modules`, `__pycache__`, anything starting with `.`, and `*.egg-info`. |
| Toolchain | `python <version>` from the venv's interpreter when `.venv` exists, else from the first system interpreter that answers: `py -3 --version`, then `python --version`, then `python3 --version` |

The edge to another project is made when an import name equals another project's export. Real output after `apps/web` imported `py_demo_util`:

```
$ smith graph
sdk  package · python · packages/sdk
util  lib · python · libs/util
web  app · python · apps/web
  └─ util
```

The toolchain probe asks `.venv`'s interpreter first, since that is the Python the tasks run with; before the venv exists it looks at the system interpreters, so `py -3` can answer 3.12.0 while `python` on PATH is 3.13.14 (see [Toolchain](#toolchain)).

## Tasks

All tasks run in the project directory with the venv interpreter, written `{workspaceRoot}/.venv/Scripts/python.exe` on Windows and `{workspaceRoot}/.venv/bin/python` elsewhere (smith expands the placeholder at run time, so the cache key is the same on every machine). Every project is installed editable, so no task depends on another project's build. Inputs of the cached tasks: `**/*`, `{workspaceRoot}/requirements-dev.txt`, the workspace-root tool configuration (`ruff.toml`, `mypy.ini`, `pytest.ini`, `pyproject.toml`, `setup.cfg`, `tox.ini`) and, when smith's default applies, that default's file under `.smith/defaults/`, so editing a default reruns the task.

| Task | Command (after `<py>`) | Appears for | Cached | Outputs |
|---|---|---|---|---|
| `lint` | `-m ruff check --config {workspaceRoot}/.smith/defaults/ruff.toml .` (no `--config` when the repository configures ruff, see [Tool configuration](#tool-configuration)) | every project | yes | none |
| `test` | `-m pytest -q -p no:cacheprovider -c {workspaceRoot}/.smith/defaults/pytest.ini --rootdir=.` (no `-c` when the repository configures pytest) | every project | yes | none |
| `typecheck` | `-m mypy --config-file {workspaceRoot}/.smith/defaults/mypy.ini .` (no `--config-file` when the repository configures mypy) | every project | yes | none |
| `build` | `-m build --outdir {workspaceRoot}/dist/<project dir> .` | `package` only | yes | `dist/<project dir>` (wheel and sdist) |
| `deps-check` | `smith py-deps-check <dir>` (not run through the venv) | `package` only; its `lint` (or, without one, `typecheck` or `test`) runs after it, so CI verify runs it | yes; inputs `**/*` and `requirements-dev.txt` | none |
| `start` | `-m <import name>` | `app` only | no | none |
| `package` | `-m build --wheel --outdir {workspaceRoot}/dist/drop .` | `app` only | no | none declared; the wheel lands in `dist/drop` for `smith ci pack` |

`deps-check` fails when a `.py` file of the package imports a module that none of the distributions its `pyproject.toml` declares provides, because the workspace's one shared venv hides the gap that whoever installs the package from PyPI hits. Declarations are read from `[project] dependencies` and `optional-dependencies` (extras count as declared), `[dependency-groups]`, Poetry's tables, `[build-system] requires`, and the project's `requirements*.txt` files; a test, `conftest.py`, `setup.py`, `noxfile.py` or a file under `tests/`, `test/`, `testing/`, `e2e/`, `tools/`, `scripts/`, `docs/`, `examples/` or `benchmarks/` may also import the development ones and the workspace's `requirements-dev.txt`. A development-only import in a file that ships is reported as `declared only in development dependencies`.

It reports only what it can be sure about. Never reported: the standard library, the project's own modules (any folder or file name inside the project), relative imports, imports in a docstring, and imports inside `try`/`except` or `if TYPE_CHECKING:` (optional by intent). A distribution's import name often differs from its name (PyYAML is imported as `yaml`, Pillow as `PIL`); smith matches an import to a distribution by name, by a table of well-known differences, or by the metadata of the distribution installed in the workspace `.venv` (`top_level.txt` or `RECORD`). When a declared distribution is none of those (not in the table, not installed, and its name is not one of the package's imports), smith cannot tell what it provides and reports nothing for that package until `smith ci setup` has installed it. A package with `dynamic = ["dependencies"]` and no `requirements.txt` is skipped for the same reason. The output lines are `<file> imports <module>, which pyproject.toml does not declare`.

A `lib` has no `build`, and an `app` has no `build` either: `smith run build` ran only the package's build in the test workspace (`1/1 tasks`). There is no `format` task.

With `[conventions.vertical-slices]` each project also gets `slice-check`, and `lint` runs after it ([vertical-slices](vertical-slices.md)).

Real output (project names from the test workspace `py-demo`):

```
$ smith explain web
web has 6 tasks:
  lint         {workspaceRoot}/.venv/Scripts/python.exe -m ruff check --config {workspaceRoot}/.smith/defaults/ruff.toml .  (cached, after slice-check)
  package      {workspaceRoot}/.venv/Scripts/python.exe -m build --wheel --outdir {workspaceRo...
  slice-check  smith slices check --lang python .  (cached)
  start        {workspaceRoot}/.venv/Scripts/python.exe -m py_demo_web
  test         {workspaceRoot}/.venv/Scripts/python.exe -m pytest -q -p no:cacheprovider -c {workspaceRoot}/.smith/defaults/pytest.ini --rootdir=.  (cached)
  typecheck    {workspaceRoot}/.venv/Scripts/python.exe -m mypy --config-file {workspaceRoot}/.smith/defaults/mypy.ini .  (cached)

$ smith explain sdk:build
sdk:build — Build the wheel and sdist into dist/packages/sdk/
  command    {workspaceRoot}/.venv/Scripts/python.exe -m build --outdir {workspaceRoot}/dist/packages/sdk .
  cwd        packages/sdk
  dependsOn  —
  inputs     **/*, {workspaceRoot}/requirements-dev.txt, {workspaceRoot}/ruff.toml, {workspaceRoot}/.ruff.toml, {workspaceRoot}/pyproject.toml
  outputs    ../../dist/packages/sdk
  env        —
  cache      true

$ smith run typecheck       # after adding [tool.mypy] strict = true to libs/util/pyproject.toml, which replaces smith's default
util:typecheck ▶ {workspaceRoot}/.venv/Scripts/python.exe -m mypy .
util:typecheck │ Success: no issues found in 5 source files
```

Running the build leaves `build/` and `<name>.egg-info` directories in the project source tree (setuptools's own); `smith new` writes `*.egg-info/` into `.gitignore`, and the build output itself goes to `dist/`.

### Tool configuration

smith owns the lint, type check and tests of a Python repository, so a new project carries no ruff, mypy or pytest configuration. Each tool has its own default, resolved the same way: the next two paragraphs describe lint, the table after them all three. When ruff finds none for a project (no `ruff.toml` or `.ruff.toml` in the project or a directory above it, up to the workspace root, and no `[tool.ruff]` table in a `pyproject.toml` there), smith writes its default to `.smith/defaults/ruff.toml` (git ignores `.smith/`; the file is rewritten whenever it differs from the one in the smith binary) and passes it with `--config`. The default is `line-length = 100`, `target-version = "py312"` and the rules `E F I UP B N`.

Your own file is the override and replaces the default whole: create `ruff.toml` at the workspace root (shared by every project) or a `[tool.ruff]` table in a project's `pyproject.toml`. When one exists smith passes no `--config` and ruff finds it, as it would without smith. To start from the default, run [`smith eject`](../../../manual/eject.md) (`smith eject ruff mypy pytest`): it copies the default to `ruff.toml`, `mypy.ini` or `pytest.ini` in the workspace root.

| Tool | Default file | Passed as | Default content | The repository's override (the nearest one up to the workspace root wins and replaces the default whole) |
|---|---|---|---|---|
| ruff | `.smith/defaults/ruff.toml` | `--config` | line length 100, `py312`, rules `E F I UP B N` | `ruff.toml`, `.ruff.toml`, `[tool.ruff]` in `pyproject.toml` |
| mypy | `.smith/defaults/mypy.ini` | `--config-file` | `python_version = 3.12`, `check_untyped_defs`, `ignore_missing_imports`, `explicit_package_bases`, warnings for redundant casts and unused ignores; no annotations required | `mypy.ini`, `.mypy.ini`, `[tool.mypy]` in `pyproject.toml`, `[mypy]` in `setup.cfg` |
| pytest | `.smith/defaults/pytest.ini` | `-c` with `--rootdir=.`, so the project stays pytest's root | `-ra --strict-markers`, `xfail_strict`, `testpaths = tests` | `pytest.ini`, `.pytest.ini`, `[tool.pytest]` or `[tool.pytest.ini_options]` in `pyproject.toml`, `[pytest]` in `tox.ini`, `[tool:pytest]` in `setup.cfg` |

Every default is written from the smith binary into `.smith/defaults/` (ignored by version control) when it differs, and its content is an input of the task, so editing it, or upgrading smith to a release that changes it, reruns the task.

## Creating a project

```
smith add app|lib|package <name> --lang python
```

The first `smith add` of a workspace also writes `requirements-dev.txt`, creates `.venv` with `<launcher> -m venv --upgrade-deps .venv` (the launcher is `py -3`, `python`, `python3` on Windows and `python3`, `python` elsewhere; the first that is on PATH and succeeds) and installs the pinned tools. Every `smith add` then installs the new project with `pip install -e ./<dir>`. This took 59 seconds on the first project (network) in the test run.

Real output:

```
$ smith add lib util --lang python
Adding lib util (python) in libs/util
Wrote 10 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  libs/util/pyproject.toml
  libs/util/src/py_demo_util/__init__.py
  libs/util/src/py_demo_util/greeting/__init__.py
  libs/util/src/py_demo_util/greeting/greet_use_case.py
  libs/util/src/py_demo_util/greeting/greeting_contract.py
  libs/util/tests/test_greet_use_case.py
  requirements-dev.txt
Import it from another Python project as `import py_demo_util`.
Try: smith test -p util

$ smith add package sdk --lang python
Wrote 8 file(s):
  .github/workflows/ci.yml
  packages/sdk/README.md
  packages/sdk/pyproject.toml
  packages/sdk/src/py_demo_sdk/__init__.py
  ...
```

An app adds `src/<import>/__main__.py`, which prints `Hello, world!`; a package adds a `README.md`.

| Generated | What it is |
|---|---|
| `pyproject.toml` | setuptools build backend (`setuptools>=75`), `name = "<workspace>-<project>"` (the workspace name is prefixed so the name is unique on a shared index; `acme` + `sdk` is `acme-sdk`), `version = "0.0.0"`, `requires-python = ">=3.12"`, empty `dependencies`, `[tool.smith] kind` (and `stack`), `[tool.setuptools.packages.find] where = ["src"]` |
| `src/<import>/greeting/` | a worked slice: `greeting_contract.py` (a frozen `Greeting` dataclass), `greet_use_case.py` (`greet(name)`), `__init__.py` re-exporting both |
| `tests/test_greet_use_case.py` | a pytest test of `greet` |
| `requirements-dev.txt` (once, at the root) | the pinned toolchain, below |

`requirements-dev.txt` as written by smith 1.62.0:

```
build==1.6.1
mypy==2.4.0
pip-audit==2.10.1
pytest==9.1.1
ruff==0.17.0
setuptools==84.2.1
twine==7.0.0
```

The plugin publishes no stack. `--stack` accepts a stack declared with `language = "python"` in `smith.toml`.

## Settings

The plugin has no `[projects.<p>.settings.python]` keys. What it reads:

| Where | Key | Effect | Read in |
|---|---|---|---|
| `pyproject.toml` | `[tool.smith] kind`, `stack` | kind and stack | `discovery/pyproject_contract.go` |
| `pyproject.toml`, `mypy.ini`, `.mypy.ini`, `setup.cfg` | `[tool.mypy]`, `[mypy]` | the repository's mypy configuration: smith passes no `--config-file` | `typechecking/mypy_args_use_case.go` |
| `pyproject.toml` | `[project] version`, `description`, `dependencies` | version, description, graph edges | same |
| `requirements-dev.txt`, `ruff.toml`, `.ruff.toml`, root `pyproject.toml` | whole file | part of every cached task's inputs; `ruff.toml`, `.ruff.toml` and a `[tool.ruff]` table are the repository's ruff override (smith creates none); edit `requirements-dev.txt` freely, smith only creates it when missing | `discovery/project_tasks_config.go` |
| `smith.toml` | `[registries.<name>] pypi`, `pypi_upload`, `auth`, `token_env` plus `[workspace] registry`, `[release] registry`, `[projects.<p>] registry` | private index for installs and for publishing | `registries/`, `releasing/upload_target_policy.go` |
| `smith.toml` | `[ci] audit_policies = ["python"]`, `audit_warn = ["python"]` | adds the requirements policy to the audit | `pipeline/ci_contribution_use_case.go` |
| `smith.toml` | `[toolchains] python = "3.12"` | pin checked by doctor | host |
| env | `PYPI_TOKEN` | PyPI API token for publishing (used as `TWINE_USERNAME=__token__` unless `TWINE_PASSWORD` is set) | `releasing/upload_target_policy.go` |
| env | `TWINE_REPOSITORY_URL` | upload to another index | same |
| env | `TWINE_USERNAME`, `TWINE_PASSWORD` | twine's own credentials (the plugin only checks `TWINE_PASSWORD` so as not to override it) | same |
| env | `SMITH_PYPI_SIMPLE_URL` | the simple index to check before uploading to a custom index | `releasing/index_client.go` |
| env | `PYPIRC_PATH` | `.pypirc` written by Azure's `TwineAuthenticate@1`; used for an `azure-artifacts` publish registry | `releasing/upload_target_policy.go` |
| env | `PIP_INDEX_URL` | the index `smith up` asks for newest versions (default `https://pypi.org/simple/`) | `dependencies/latest_versions_use_case.go` |

The tool versions in `requirements-dev.txt` and the Python version `3.12` are constants of the plugin (`venv/dev_tools_config.go`); smith does not rewrite an existing `requirements-dev.txt`.

## Toolchain

- One virtual environment per workspace, `.venv` at the root. Everything, including the pinned tools, lives there; there is no per-project environment.
- The interpreter that creates it must exist on the machine: `py -3` / `python` / `python3` on Windows, `python3` / `python` elsewhere. When none is found: `no Python found on PATH (tried ...); install Python 3.12 or newer`.
- `[toolchains] python = "..."` is compared with the version of the venv's interpreter, or without a venv the first system interpreter that answers `--version` (the probe order is in the table above). Real output, on a machine without a venv, with a pin of `3.13` on a machine where `py -3` is 3.12.0:

```
✖ python matches the pinned version 3.13
    [toolchains] pins python 3.13, but this machine runs 3.12.0: tasks cached on another version are misses, and results can differ from CI
    → install python 3.13 (`mise use python@3.13`, asdf, or your installer; `smith doctor --fix` writes .tool-versions for mise and asdf), or change [toolchains] python
```

- Doctor separately warns when the venv's own Python is not `3.12`, the version CI installs; see below.
- Output: `build` writes `dist/<project dir>/*.whl` and `*.tar.gz`; `package` writes `dist/drop/<dist>-<version>-py3-none-any.whl` (real: `py_demo_web-0.0.0-py3-none-any.whl`).

## Releasing

| Aspect | Behaviour |
|---|---|
| Version | `[project] version` in `pyproject.toml`. The release writes the new version into the first `version = "..."` line of the `[project]` table (quoting, spacing, comments, line endings kept), builds, publishes, then restores the file: the tag is the record. With `--dry-run` nothing is written. A `pyproject.toml` with no `version` in `[project]` fails: `pyproject.toml has no version in [project]`. |
| Tag | `<name>@<version>`, name being the directory name: `sdk@0.1.0` (real dry-run output). Override with `tag_prefix`. |
| What is published | `package` projects only. Apps are never uploaded; an app's wheel can be sent with `[projects.<app>.release] pypi = "..."` (see [releasing](../../../reference/releasing.md#apps)). |
| How | `twine upload --non-interactive --disable-progress-bar <files>` run from the project directory, for the files in `dist/<project dir>` whose name contains `-<version>`. The build must exist (`no distribution for version ... ; build it first`). |
| Skipping | before uploading, the plugin reads the simple index (`<index>/<normalised name>/`) and uploads only files it lacks. If twine itself reports `File already exists` or `409`, the result is "already published". An unreachable index reports nothing, so the upload is attempted. |
| Registry | PyPI (`https://upload.pypi.org/legacy/`) unless `[registries.*] pypi_upload` is set for the project, or `TWINE_REPOSITORY_URL` is set. The simple index for the existence check is derived (`upload.pypi.org` to `pypi.org/simple/`, `test.pypi.org` likewise, `.../upload/` to `.../simple/`, anything else `<url>/simple/`). |
| Credentials | `PYPI_TOKEN` (listed as a required secret when a package publishes to PyPI itself), or `TWINE_USERNAME` / `TWINE_PASSWORD`; for a `token` registry the variable named in `token_env`; for Azure Artifacts none (the pipeline's own token through `TwineAuthenticate@1`; elsewhere twine signs in through `artifacts-keyring`). |
| Release channel | the plugin declares `channels`: PyPI orders a prerelease (`2.0.0b1`) below its release, so a channel version never becomes the one `pip install` picks, and no tag is needed |
| Dry run | publish returns "dry run" and uploads nothing. Real output: |

```
$ smith release --dry-run
Dry run: nothing is published or tagged.
  sdk                  unreleased → 0.1.0  (1 releasing commit(s))

$ smith release --dry-run --json      # excerpt
  "tag": "sdk@0.1.0",  "publish": "dry run",  "registry": "https://upload.pypi.org/legacy/"
```

The `published` capability lets smith ask the index for the newest version it has (the plugin reads the simple index page); which host command consumes it was not traced.

The release's SBOM lists the packages of `uv.lock` or `poetry.lock` (the project's or any directory above), falling back to `requirements.txt` and the dependencies of `pyproject.toml`, through the `sbom/components` method of the plugin protocol ([releasing](../../../reference/releasing.md#sbom-and-build-provenance)).

## CI

Real output of `smith ci generate` for the test workspace (GitHub):

```yaml
      - uses: actions/setup-python@v7
        with:
          python-version: '3.12'
      - name: Cache the pip download store
        uses: actions/cache@v4
        with:
          path: ~/.cache/pip
          key: smith-pip-${{ runner.os }}-${{ hashFiles('requirements-dev.txt', 'packages/sdk/pyproject.toml', 'libs/util/pyproject.toml', 'apps/web/pyproject.toml') }}
      ...
      - name: Release
        if: github.event_name == 'push' && github.ref_name == 'main'
        run: smith ci release
        env:
          PYPI_TOKEN: ${{ secrets.PYPI_TOKEN }}
```

| Contribution | What |
|---|---|
| Toolchain | Python `3.12`: `actions/setup-python@v7` (GitHub), `UsePythonVersion@0` (Azure), `uv python install 3.12` plus symlinks for other providers, the `python` dev container feature at 3.12 |
| Package store | `~/.cache/pip` (`PIP_CACHE_DIR`), keyed on `requirements-dev.txt` and each project's `pyproject.toml` |
| Setup (`smith ci setup`) | `Create the Python venv` (`<launcher> -m venv --upgrade-deps .venv`), `Show the Python version` (`python --version`, non-blocking), then, if a private index is configured, `Point pip at the private index` (`pip config --site set global.index-url <index>`), `Install the Python toolchain` (`pip install -r requirements-dev.txt`), and `Install <project>` (`pip install -e ./<root>`, without `--quiet` so a resolver conflict is printed in full) per project. All blocking except the version. Real: five steps, `smith ci setup passed in 58.9s`. |
| Audit | `pip-audit --skip-editable`, non-blocking, cached on `requirements-dev.txt` and the `pyproject.toml` files. With `[ci] audit_policies = ["python"]` also `smith py-policy` over `requirements-dev.txt` and each project's `requirements.txt`: every requirement in a project's `requirements.txt` must have a `--hash` (the generated `requirements-dev.txt` is exempt: it lists top-level pins only, and hash-checking mode needs a hash for every transitive dependency), none may be a URL or git reference, and the only index is the workspace's. |
| Credentials | `PYPI_TOKEN` (required; only when a `package` publishes to PyPI itself). `smith ci check` printed `GitHub: add the repository secrets PYPI_TOKEN` and the Azure equivalent. |
| Private index (Azure Artifacts) | `PipAuthenticate@1` (`artifactFeeds: 'libs'`) for installs and `TwineAuthenticate@1` (`artifactFeed: 'libs'`) for publishing, no secret. Real, with `[registries.libs] auth = "azure-artifacts"` and `[workspace] registry = "libs"`; `smith ci check` then asked for the Feed Publisher role on the build service. A `token` registry adds the variable named in `token_env` as a secret (given to setup when installs need it). |

Real audit output (before `setuptools` moved from 84.0.0 to 84.2.1 the pinned toolchain itself was flagged):

```
▶ pip-audit
No known vulnerabilities found
smith ci audit passed
```

With the `python` audit policy turned on, the generated `requirements-dev.txt` passes: the hash rule applies to each project's `requirements.txt`, while the index and URL rules still cover the toolchain file.

## Linking an internal library

`smith install -w <project> shop-core-py` writes the library's distribution name (the `[project] name` of its `pyproject.toml`, not its import name) in `[project] dependencies`. The library is installed editable in the shared `.venv` before the project, so pip finds the requirement there and never asks PyPI; `smith install` and `smith ci` install projects in dependency order for the same reason. `smith uninstall` removes the line and leaves the library in the venv. The plugin declares the `link` capability.

## Doctor, dependencies and removal

Doctor checks (`pipeline/ci_contribution_use_case.go`, `check_python_version_use_case.go`; shown only when the workspace has Python projects):

| Check | Fails when | Remedy printed |
|---|---|---|
| `Python venv exists` | `.venv/Scripts/python.exe` (or `.venv/bin/python`) is missing | run `smith ci setup` |
| `requirements-dev.txt exists` | the file is missing | run `smith add` for a Python project, or restore it |
| `<dir> declares the distributions it imports` | a package (`kind = package`) fails `deps-check`; no automatic fix | add the distribution to `dependencies` in `<dir>/pyproject.toml` |
| `the venv's Python is 3.12, as in CI` | a warning when the venv's Python is newer than `3.12` (or has none); a failure when it is older | delete `.venv`, then `smith ci setup` |

All three passed in the test run (`17 passed, 0 failed, 0 warning(s)`).

Dependencies:

| Command | What it does |
|---|---|
| `smith install` | makes `.venv` and `requirements-dev.txt` if missing, installs the tools, installs every project editable |
| `smith install -w <project> <req>...` | adds the requirement(s) to `[project] dependencies` in that project's `pyproject.toml` (a TOML edit that keeps the rest of the file), then `pip install -e ./<dir>` |
| `smith install -w <project> -D <req>` | adds to `[project.optional-dependencies] dev`, then `pip install <req>` by name |
| `smith uninstall -w <project> <name>` | removes the requirement from the runtime list and the `dev` list. The package stays installed in the shared `.venv`; the note says to recreate it. Fails with `<project> does not require <name>` if neither list has it. |
| `smith up`, `smith sync` | the plugin lists `[project] dependencies` and the `dev` extra (`declarations`) and rewrites a plain constraint (`==`, `>=`, `~=`, `<=`, `<`, `>` with a version); compound ranges, `!=`, markers, URLs are reported and left alone. Newest versions come from the simple index (`PIP_INDEX_URL`, default PyPI); `smith up` also compares the pinned Python with the newest series (from `endoflife.date`). |

Real output:

```
$ smith install -w sdk "requests>=2.0"
✔ sdk: added requests>=2.0 (packages/sdk/pyproject.toml)
$ smith install -w sdk -D pytest-cov
✔ sdk: added pytest-cov (packages/sdk/pyproject.toml)
$ smith up --check
minor
  requests                     2.0 → 2.34.2  sdk

Tools smith pins (informational):
  Python           3.12 → 3.15 available
$ smith uninstall -w sdk requests
✔ sdk: removed requests (packages/sdk/pyproject.toml)
  left installed in .venv, which every project shares; recreate it to drop requests
```

Removal: `smith remove <project>` deletes the directory and runs `pip uninstall --yes <import name>` in `.venv` so nothing imports a package whose source is gone. Real dry run: `would uninstall py_demo_web from .venv`.

## Declared projects

A folder with no `pyproject.toml` of its own, under a `pyproject.toml` with a `[project]` table (the nearest one above it), can be declared with `[projects.<name>]`, a `root` and `language = "python"`. Its import name, the folder name with `-` and `.` turned into `_`, is the export, so a project that imports it depends on it. Its tasks are `lint` (ruff), `typecheck` (mypy) and `test` (pytest) run in the folder, and an app also has `start` (`python -m <import name>`). The `pyproject.toml` above is an input of each task. A declared project is an app or a lib, never a package: the distribution has one version. A missing folder, a folder with a `pyproject.toml` of its own and a missing manifest above are refused with a message that names the `[projects.<name>]` table.

## Requirements and limits

- A Python 3.12 or newer on the machine (`requires-python = ">=3.12"` in generated projects; CI installs exactly 3.12). The venv is created from whatever the launcher finds first.
- Network: creating the venv and installing the tools needs the index (the first `smith add` took about a minute here).
- One environment for all projects: two projects that need conflicting versions of a dependency cannot both be installed. `smith uninstall` never removes a package from the venv.
- The editable install makes a project's own `.egg-info` appear under `src/`; it is git-ignored by `smith new`.
- Only `pyproject.toml` projects are discovered. `setup.py`-only and `requirements.txt`-only directories are not, and a `pyproject.toml` that has no `[project]` table is skipped silently.
- There is no `format` task and the apps have no `build`.

## Verified

Windows 11 Pro 10.0.26300, smith 1.62.0, Python 3.12.0 (`py -3`; `.venv` created with it; the `python` on PATH is 3.13.14, which the plugin does not use), pip-installed tools as pinned above (ruff, pytest, mypy, build, twine, pip-audit, setuptools). Throw-away workspace `py-demo` from `smith new py-demo`. Run: `smith add lib|app|package ... --lang python`, `smith projects`, `smith graph`, `smith explain`, `smith run build|test|lint|typecheck|package|start`, `smith doctor` (clean, and with a wrong `[toolchains] python` pin), `smith ci generate` (GitHub; then GitHub and Azure with an Azure Artifacts registry), `smith ci setup`, `smith ci audit` (with and without the `python` audit policy), `smith ci check`, `smith release --dry-run [--json]`, `smith install -w` (runtime and `-D`), `smith uninstall -w`, `smith up --check`, `smith remove web --dry-run`, `smith plugin check python` (11 checks ok).

Not verified:

- A real upload: `smith release` without `--dry-run` would publish to PyPI. The twine invocation, the already-published skip and the token wiring are read from `releasing/`.
- `smith ci setup` with a private index (it would write a pip site configuration pointing at a host that does not exist); only the generated pipelines were inspected.
- `PipAuthenticate@1` and `TwineAuthenticate@1` on a real Azure agent; GitLab output; the Linux and macOS venv paths (`.venv/bin/python`).
- `smith sync` / `smith up --align` rewriting Python constraints (the `declarations` code path was not driven).

## See also

- [The plugin catalogue](README.md)
- [vertical-slices](vertical-slices.md): the `slice-check` task Python projects get
- [Releasing](../../../reference/releasing.md), [CI](../../../reference/ci.md), [Configuration: private registries](../../../reference/configuration.md#private-registries)
- [Writing a language plugin](../../creating/language-plugin.md)
