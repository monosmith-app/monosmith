package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.trycommand.TryCommandUseCase
import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase

/** Writes the smith launcher (`smith bootstrap`) after its plan. */
class BootstrapLauncherActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.writeLauncher(ide, root) })

/** Installs the MCP server and the agent files for Claude Code, after the diffs. */
class InstallAgentActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.installAgent(ide, root) })

/** Tries a smith command in a copy of the workspace, then offers to run it for real. */
class TryCommandActionHandler : SmithActionHandler({ ide, root -> TryCommandUseCase.start(ide, root) })

/** Explains how the tasks of a project resolved (`smith explain`). */
class ExplainProjectActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.explain(ide, root, null) })
