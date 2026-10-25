"""Building the answers of `ci`."""

from __future__ import annotations

from collections.abc import Mapping, Sequence

from ..protocol import ActionStep, Credential, Step


def action_step(uses: str, inputs: Mapping[str, str] | Sequence[tuple[str, str]] | None = None) -> ActionStep:
    """A provider-native step: a GitHub action (`"actions/setup-node@v6"`) or an Azure task (`"NodeTool@0"`).

    The inputs are written in the order given. The values of a pipeline are written as quoted YAML
    strings, except `true` and `false`, so pass strings.
    """
    pairs = list(inputs.items()) if isinstance(inputs, Mapping) else list(inputs or [])
    if not pairs:
        return {"uses": uses}
    return {"uses": uses, "inputs": [{"name": name, "value": value} for name, value in pairs]}


def ci_step(
    name: str,
    command: Sequence[str],
    *,
    dir: str | None = None,  # noqa: A002 - the protocol's member name
    blocking: bool = True,
    cache_files: Sequence[str] | None = None,
) -> Step:
    """One command a CI phase (`setup`, `audit`, `postRelease`) runs. It blocks unless `blocking=False`.

    `dir` is the workspace-relative directory to run in (default: the workspace root). `cache_files`
    are workspace-relative files whose content decides what an audit step finds (a lockfile): smith
    skips a passing step while they are unchanged.
    """
    step: Step = {"name": name, "command": list(command), "blocking": blocking}
    if dir is not None:
        step["dir"] = dir
    if cache_files is not None:
        step["cacheFiles"] = list(cache_files)
    return step


def credential(
    env: str, description: str, *, required: bool = True, setup: bool = False, tasks: bool = False
) -> Credential:
    """A secret the release phase needs, read from the environment variable `env`.

    `required=False` says release works without it. `setup` also gives it to `smith ci setup` (installs
    from a private registry need it); `tasks` also gives it to every step that runs tasks.
    """
    result: Credential = {"env": env, "description": description, "required": required}
    if setup:
        result["setup"] = True
    if tasks:
        result["tasks"] = True
    return result
