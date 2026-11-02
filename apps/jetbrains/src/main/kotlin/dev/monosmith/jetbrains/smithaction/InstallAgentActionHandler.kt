package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase

/** Installs the MCP server and the agent files for Claude Code, after the diffs. */
class InstallAgentActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.installAgent(ide, root) })
