"""Building the tasks a plugin contributes or discovers."""

from __future__ import annotations

from collections.abc import Callable, Mapping, Sequence
from typing import Any

from ..build_dir import build_dir_from_project
from ..protocol import Task


def task(
    command: Sequence[str],
    *,
    cwd: str | None = None,
    env: Mapping[str, str] | None = None,
    inputs: Sequence[str] | None = None,
    outputs: Sequence[str] | None = None,
    depends_on: Sequence[str] | None = None,
    cache: bool | None = None,
    description: str | None = None,
    verify: bool | None = None,
    serve: str | None = None,
    serve_url: str | None = None,
    serve_timeout: str | None = None,
    **extra: Any,
) -> Task:
    """A task the host can run, hash and cache.

    The command is an argv list and is never passed through a shell. The keyword arguments are the
    task's other members, left out when `None` (`depends_on` is the protocol's `dependsOn`); `extra`
    passes members this SDK does not know yet, under the protocol's own names.

        task(["python", "check.py"], inputs=["widget.json"], cache=True, description="check the manifest")
    """
    named: dict[str, Any] = {
        "cwd": cwd,
        "env": None if env is None else dict(env),
        "inputs": None if inputs is None else list(inputs),
        "outputs": None if outputs is None else list(outputs),
        "dependsOn": None if depends_on is None else list(depends_on),
        "cache": cache,
        "description": description,
        "verify": verify,
        "serve": serve,
        "serveUrl": serve_url,
        "serveTimeout": serve_timeout,
    }
    result: dict[str, Any] = {"command": list(command)}
    result.update({key: value for key, value in named.items() if value is not None})
    result.update(extra)
    return result  # type: ignore[return-value]


def build_task(
    root: str,
    command: Callable[[str], Sequence[str]],
    **options: Any,
) -> Task:
    """The `build` task of a project: cached, built into the project's directory under `dist/`.

    That directory is also its `outputs`, restored on a cache hit, and the task runs after the build of
    every project this one imports (`^build`). `command` is called with the directory the command is
    told to write into, `../../dist/libs/core` for `libs/core`, to put in its argv. The keyword
    arguments are those of `task`, and win over the defaults (`cache=True`, `depends_on=["^build"]`).

        build_task(root, lambda out: ["python", "build.py", out], inputs=["widget.json"])
    """
    output_dir = build_dir_from_project(root)
    defaults: dict[str, Any] = {"outputs": [output_dir], "depends_on": ["^build"], "cache": True}
    return task(command(output_dir), **{**defaults, **options})
