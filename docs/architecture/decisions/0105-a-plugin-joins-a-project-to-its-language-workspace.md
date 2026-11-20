# ADR 0105: A plugin joins a project to its language's workspace

- Status: accepted
- Date: 2026-11-20
- Issue: #1151 (epic #1156, findings F2 to F5 of
  [the review](../research/any-repo-into-a-workspace-review.md))

## Context

smith sees a project when a plugin discovers its manifest. That does not make the project part of
its language's workspace. A Node project that the root package.json "workspaces" do not cover is
listed by `smith projects`, but `npm install` does not reach it. Its dependencies are missing, and
`smith install` still said "installed 2 project(s)". Doctor reported the gap with no fix. A project
that someone copies into the workspace, and the coming `smith import` (#1152), need one operation
that joins the project, for every language.

## Decision

1. The plugin protocol gets the method `workspace/join` and the capability `join`. The host sends
   the plugin's projects (`JoinParams{projects}`). The plugin answers one `JoinState` for each
   project its language's workspace is about: `project`, `joined`, `workspace` (what it joins, in
   words), `problem` and `fix`. A project the plugin leaves out is joined. The plugin writes nothing:
   the fix is the same `Fix` that doctor findings carry (files, then steps, then choices).
2. Node: the root package.json "workspaces" must match the project folder. The fix adds the parent
   folder's glob (`apps/*`), or the folder for a project at the top, keeps the rest of the manifest
   as written, then runs `npm install`. A root without package.json gets a private one. The object
   form (`{"packages": [...]}`) gets no fix.
3. Python: only when the root pyproject.toml has `[tool.uv.workspace] members`. The fix adds the
   parent folder's glob, the change `smith kind` already makes when a project moves. Without a uv
   workspace, smith installs every project into the workspace venv, so nothing is missing.
4. .NET: only when the workspace root has a solution (`.sln`, `.slnx`). The fix is
   `dotnet sln <solution> add <project file>`. With several solutions the fix asks which one.
   Without a solution, smith restores every project on its own.
5. Doctor asks every plugin with the capability. Each project gets the check
   "<folder> is joined to its language's workspace", with the plugin's fix and, in the remedy, the
   way to keep the project out on purpose (`[workspace] exclude`). The Node plugin's own check
   "npm workspaces include <folder>", which had no fix, is removed.
6. `DependencyResult` gets `unreached`: the projects an install of every project did not reach.
   `smith install` leaves them out of its count, names them with the step that joins them, and exits
   with an error.

`smith add` does not join a project for Python or .NET: it makes no uv workspace and no solution.
So the join for those languages applies to repositories that bring their own.

## Consequences

- `smith doctor --fix` joins a project that was copied in by hand. `smith import` calls the same
  method after it places a project.
- PHP (a root composer.json path repository), Go (`go.work use`), Rust (Cargo members) and Java
  (Maven modules) join later through the same method.
- Python and .NET install counts stay as they were: those plugins reach every project.
